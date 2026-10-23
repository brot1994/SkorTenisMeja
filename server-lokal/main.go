package main

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

//go:embed index.html
var indexHTML []byte

type Hub struct {
	mu         sync.Mutex
	state      map[string]any
	activeID   string
	version    int64
	pin        string
	dataDir    string
	tDir       string
	subs       map[chan []byte]bool
	tunnel     string
	tunnelStat string
	port       int
	lastBackup time.Time
}

var validID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,90}$`)

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	dash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash && b.Len() > 0 {
			b.WriteRune('-')
			dash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 40 {
		out = strings.Trim(out[:40], "-")
	}
	if out == "" {
		out = "turnamen"
	}
	return out
}

func (h *Hub) snapshot() []byte {
	out := map[string]any{}
	for k, v := range h.state {
		out[k] = v
	}
	out["id"] = h.activeID
	out["version"] = h.version
	out["serverTime"] = time.Now().UnixMilli()
	out["pinSet"] = h.pin != ""
	b, _ := json.Marshal(out)
	return b
}

func writeAtomic(path string, b []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	os.Remove(path)
	return os.Rename(tmp, path)
}

func (h *Hub) save() {
	if h.activeID == "" {
		return
	}
	h.state["id"] = h.activeID
	h.state["version"] = h.version
	b, _ := json.MarshalIndent(h.state, "", " ")
	writeAtomic(filepath.Join(h.tDir, h.activeID+".json"), b)
	if time.Since(h.lastBackup) > 5*time.Minute {
		h.lastBackup = time.Now()
		bd := filepath.Join(h.dataDir, "backup", h.activeID)
		os.MkdirAll(bd, 0755)
		os.WriteFile(filepath.Join(bd, time.Now().Format("20060102-150405")+".json"), b, 0644)
		files, _ := filepath.Glob(filepath.Join(bd, "*.json"))
		sort.Strings(files)
		for len(files) > 60 {
			os.Remove(files[0])
			files = files[1:]
		}
	}
}

func (h *Hub) setActive(id string) {
	h.activeID = id
	os.WriteFile(filepath.Join(h.dataDir, "aktif.txt"), []byte(id), 0644)
}

func (h *Hub) broadcast() {
	b := h.snapshot()
	for c := range h.subs {
		select {
		case c <- b:
		default:
		}
	}
}

func (h *Hub) checkPin(w http.ResponseWriter, r *http.Request) bool {
	if h.pin != "" && r.Header.Get("X-PIN") != h.pin {
		w.WriteHeader(401)
		w.Write([]byte(`{"error":"pin"}`))
		return false
	}
	return true
}

func (h *Hub) load() {
	os.MkdirAll(h.tDir, 0755)
	// migrasi data versi lama (satu turnamen)
	old := filepath.Join(h.dataDir, "turnamen.json")
	if b, err := os.ReadFile(old); err == nil {
		id := "turnamen-lama-" + time.Now().Format("20060102")
		var st map[string]any
		if json.Unmarshal(b, &st) == nil {
			st["id"] = id
			nb, _ := json.MarshalIndent(st, "", " ")
			writeAtomic(filepath.Join(h.tDir, id+".json"), nb)
			os.Rename(old, old+".lama")
			if _, err := os.Stat(filepath.Join(h.dataDir, "aktif.txt")); err != nil {
				h.setActive(id)
			}
		}
	}
	if b, err := os.ReadFile(filepath.Join(h.dataDir, "aktif.txt")); err == nil {
		id := strings.TrimSpace(string(b))
		if validID.MatchString(id) {
			if fb, err := os.ReadFile(filepath.Join(h.tDir, id+".json")); err == nil {
				var st map[string]any
				if json.Unmarshal(fb, &st) == nil {
					h.state = st
					h.activeID = id
					if v, ok := st["version"].(float64); ok {
						h.version = int64(v)
					}
				}
			}
		}
	}
	if b, err := os.ReadFile(filepath.Join(h.dataDir, "pin.txt")); err == nil {
		h.pin = strings.TrimSpace(string(b))
	}
}

func noCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

var extraIPs []string

func lanURLs(port int) []string {
	var out []string
	seen := map[string]bool{}
	for _, ip := range extraIPs {
		ip = strings.TrimSpace(ip)
		if ip != "" && !seen[ip] {
			seen[ip] = true
			out = append(out, fmt.Sprintf("http://%s:%d/", ip, port))
		}
	}
	ifaces, _ := net.Interfaces()
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok || ipn.IP.To4() == nil {
				continue
			}
			ip := ipn.IP.String()
			if strings.HasPrefix(ip, "169.254.") || seen[ip] {
				continue
			}
			seen[ip] = true
			out = append(out, fmt.Sprintf("http://%s:%d/", ip, port))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		pi := strings.Contains(out[i], "://192.168.")
		pj := strings.Contains(out[j], "://192.168.")
		if pi != pj {
			return pi
		}
		return out[i] < out[j]
	})
	return out
}

func main() {
	port := flag.Int("port", 8080, "port web")
	internet := flag.Bool("internet", true, "buat alamat internet otomatis (butuh cloudflared.exe di folder yang sama)")
	noBrowser := flag.Bool("nobrowser", false, "jangan buka browser otomatis")
	dataFlag := flag.String("data", "", "folder data (default: folder program/data)")
	ipsFlag := flag.String("ips", "", "daftar IP tambahan untuk ditampilkan, dipisah koma")
	flag.Parse()
	if *ipsFlag != "" {
		extraIPs = strings.Split(*ipsFlag, ",")
	}

	exe, _ := os.Executable()
	baseDir := filepath.Dir(exe)
	if strings.Contains(baseDir, "go-build") {
		baseDir, _ = os.Getwd()
	}
	h := &Hub{state: map[string]any{}, subs: map[chan []byte]bool{}, port: *port}
	h.dataDir = filepath.Join(baseDir, "data")
	if *dataFlag != "" {
		h.dataDir = *dataFlag
	}
	h.tDir = filepath.Join(h.dataDir, "turnamen")
	h.load()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			http.NotFound(w, r)
			return
		}
		noCache(w)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		noCache(w)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			h.mu.Lock()
			b := h.snapshot()
			h.mu.Unlock()
			w.Write(b)
			return
		}
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 20<<20))
		if err != nil {
			w.WriteHeader(400)
			return
		}
		var req struct {
			Base  int64          `json:"base"`
			ID    string         `json:"id"`
			State map[string]any `json:"state"`
		}
		if json.Unmarshal(body, &req) != nil || req.State == nil {
			w.WriteHeader(400)
			return
		}
		h.mu.Lock()
		defer h.mu.Unlock()
		if !h.checkPin(w, r) {
			return
		}
		if h.activeID == "" || req.ID != h.activeID || req.Base != h.version {
			w.WriteHeader(409)
			w.Write(h.snapshot())
			return
		}
		delete(req.State, "serverTime")
		delete(req.State, "pinSet")
		h.version++
		h.state = req.State
		h.save()
		h.broadcast()
		w.Write(h.snapshot())
	})
	mux.HandleFunc("/api/list", func(w http.ResponseWriter, r *http.Request) {
		noCache(w)
		w.Header().Set("Content-Type", "application/json")
		h.mu.Lock()
		active := h.activeID
		h.mu.Unlock()
		files, _ := filepath.Glob(filepath.Join(h.tDir, "*.json"))
		out := []map[string]any{}
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				continue
			}
			var st struct {
				Event struct {
					Title string `json:"title"`
					Place string `json:"place"`
					Dates string `json:"dates"`
				} `json:"event"`
				Status  string `json:"status"`
				Created int64  `json:"created"`
				Cats    []struct {
					Name string `json:"name"`
				} `json:"cats"`
			}
			if json.Unmarshal(b, &st) != nil {
				continue
			}
			id := strings.TrimSuffix(filepath.Base(f), ".json")
			cats := []string{}
			for _, c := range st.Cats {
				cats = append(cats, c.Name)
			}
			upd := st.Created
			if fi, err := os.Stat(f); err == nil && upd == 0 {
				upd = fi.ModTime().UnixMilli()
			}
			out = append(out, map[string]any{"id": id, "title": st.Event.Title, "place": st.Event.Place, "dates": st.Event.Dates,
				"status": st.Status, "active": id == active, "cats": cats, "updated": upd})
		}
		b, _ := json.Marshal(out)
		w.Write(b)
	})
	mux.HandleFunc("/api/get", func(w http.ResponseWriter, r *http.Request) {
		noCache(w)
		w.Header().Set("Content-Type", "application/json")
		id := r.URL.Query().Get("id")
		if !validID.MatchString(id) {
			w.WriteHeader(400)
			return
		}
		if r.URL.Query().Get("dl") == "1" {
			w.Header().Set("Content-Disposition", "attachment; filename=\""+id+"-"+time.Now().Format("20060102-1504")+".json\"")
		}
		h.mu.Lock()
		if id == h.activeID {
			b := h.snapshot()
			h.mu.Unlock()
			w.Write(b)
			return
		}
		h.mu.Unlock()
		b, err := os.ReadFile(filepath.Join(h.tDir, id+".json"))
		if err != nil {
			w.WriteHeader(404)
			return
		}
		w.Write(b)
	})
	mux.HandleFunc("/api/create", func(w http.ResponseWriter, r *http.Request) {
		noCache(w)
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		var req struct {
			State map[string]any `json:"state"`
		}
		if json.NewDecoder(io.LimitReader(r.Body, 20<<20)).Decode(&req) != nil || req.State == nil {
			w.WriteHeader(400)
			return
		}
		h.mu.Lock()
		defer h.mu.Unlock()
		if !h.checkPin(w, r) {
			return
		}
		title := ""
		if ev, ok := req.State["event"].(map[string]any); ok {
			title, _ = ev["title"].(string)
		}
		id := slugify(title) + "-" + time.Now().Format("20060102-150405")
		h.save()
		delete(req.State, "serverTime")
		delete(req.State, "pinSet")
		h.version++
		h.state = req.State
		h.setActive(id)
		h.lastBackup = time.Time{}
		h.save()
		h.broadcast()
		w.Write(h.snapshot())
	})
	mux.HandleFunc("/api/activate", func(w http.ResponseWriter, r *http.Request) {
		noCache(w)
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			ID string `json:"id"`
		}
		if r.Method != http.MethodPost || json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req) != nil || !validID.MatchString(req.ID) {
			w.WriteHeader(400)
			return
		}
		h.mu.Lock()
		defer h.mu.Unlock()
		if !h.checkPin(w, r) {
			return
		}
		b, err := os.ReadFile(filepath.Join(h.tDir, req.ID+".json"))
		var st map[string]any
		if err != nil || json.Unmarshal(b, &st) != nil {
			w.WriteHeader(404)
			return
		}
		h.save()
		h.version++
		h.state = st
		h.setActive(req.ID)
		h.save()
		h.broadcast()
		w.Write(h.snapshot())
	})
	mux.HandleFunc("/api/delete", func(w http.ResponseWriter, r *http.Request) {
		noCache(w)
		var req struct {
			ID string `json:"id"`
		}
		if r.Method != http.MethodPost || json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req) != nil || !validID.MatchString(req.ID) {
			w.WriteHeader(400)
			return
		}
		h.mu.Lock()
		defer h.mu.Unlock()
		if !h.checkPin(w, r) {
			return
		}
		if req.ID == h.activeID {
			w.WriteHeader(409)
			return
		}
		td := filepath.Join(h.tDir, "_terhapus")
		os.MkdirAll(td, 0755)
		if err := os.Rename(filepath.Join(h.tDir, req.ID+".json"), filepath.Join(td, req.ID+".json")); err != nil {
			w.WriteHeader(404)
			return
		}
		w.Write([]byte(`{"ok":true}`))
	})
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		fl, ok := w.(http.Flusher)
		if !ok {
			w.WriteHeader(500)
			return
		}
		noCache(w)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		ch := make(chan []byte, 8)
		h.mu.Lock()
		h.subs[ch] = true
		first := h.snapshot()
		h.mu.Unlock()
		defer func() { h.mu.Lock(); delete(h.subs, ch); h.mu.Unlock() }()
		fmt.Fprintf(w, "retry: 2000\ndata: %s\n\n", first)
		fl.Flush()
		tick := time.NewTicker(15 * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-r.Context().Done():
				return
			case b := <-ch:
				fmt.Fprintf(w, "data: %s\n\n", b)
				fl.Flush()
			case <-tick.C:
				fmt.Fprint(w, ": ping\n\n")
				fl.Flush()
			}
		}
	})
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		noCache(w)
		w.Header().Set("Content-Type", "application/json")
		h.mu.Lock()
		b, _ := json.Marshal(map[string]any{"lan": lanURLs(h.port), "internet": h.tunnel, "tunnelStatus": h.tunnelStat})
		h.mu.Unlock()
		w.Write(b)
	})
	mux.HandleFunc("/api/pin", func(w http.ResponseWriter, r *http.Request) {
		noCache(w)
		if r.Method != http.MethodPost {
			w.WriteHeader(405)
			return
		}
		var req struct {
			Pin string `json:"pin"`
		}
		json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&req)
		h.mu.Lock()
		defer h.mu.Unlock()
		if !h.checkPin(w, r) {
			return
		}
		h.pin = strings.TrimSpace(req.Pin)
		os.WriteFile(filepath.Join(h.dataDir, "pin.txt"), []byte(h.pin), 0600)
		h.broadcast()
		w.Write([]byte(`{"ok":true}`))
	})

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		fmt.Printf("\nGAGAL membuka port %d: %v\nMungkin turnamen.exe sudah berjalan di jendela lain.\n", *port, err)
		fmt.Println("Tekan Enter untuk menutup...")
		bufio.NewReader(os.Stdin).ReadString('\n')
		os.Exit(1)
	}

	fmt.Println("==============================================")
	fmt.Println("  SKOR TENIS MEJA - server turnamen")
	fmt.Println("==============================================")
	fmt.Println("Data disimpan di:", h.dataDir)
	fmt.Println()
	fmt.Println("Buka di TV / tablet yang tersambung ke WiFi yang sama:")
	for _, u := range lanURLs(*port) {
		fmt.Println("   TV    :", u+"#tv")
		fmt.Println("   Meja 1:", u+"#meja1")
		fmt.Println("   Panitia:", u+"#admin")
		fmt.Println()
	}
	fmt.Println("JANGAN tutup jendela ini selama turnamen berlangsung.")
	fmt.Println()

	if *internet {
		cf := filepath.Join(baseDir, "cloudflared.exe")
		if runtime.GOOS != "windows" {
			cf = filepath.Join(baseDir, "cloudflared")
		}
		if _, err := os.Stat(cf); err == nil {
			h.tunnelStat = "sedang dibuat..."
			go runTunnel(h, cf, *port)
		} else {
			h.tunnelStat = "cloudflared.exe tidak ditemukan"
		}
	} else {
		h.tunnelStat = "dimatikan (mode WiFi saja)"
	}

	if !*noBrowser && runtime.GOOS == "windows" {
		go func() {
			time.Sleep(800 * time.Millisecond)
			exec.Command("cmd", "/c", "start", "", fmt.Sprintf("http://localhost:%d/#admin", *port)).Start()
		}()
	}

	srv := &http.Server{Handler: mux}
	srv.Serve(ln)
}

func runTunnel(h *Hub, cf string, port int) {
	re := regexp.MustCompile(`https://[a-z0-9-]+\.trycloudflare\.com`)
	for attempt := 1; ; attempt++ {
		cmd := exec.Command(cf, "tunnel", "--no-autoupdate", "--url", fmt.Sprintf("http://localhost:%d", port))
		hideWindow(cmd)
		stderr, _ := cmd.StderrPipe()
		stdout, _ := cmd.StdoutPipe()
		if err := cmd.Start(); err != nil {
			h.mu.Lock()
			h.tunnelStat = "gagal menjalankan cloudflared: " + err.Error()
			h.mu.Unlock()
			return
		}
		scan := func(rd io.Reader) {
			sc := bufio.NewScanner(rd)
			for sc.Scan() {
				if u := re.FindString(sc.Text()); u != "" {
					h.mu.Lock()
					if h.tunnel != u+"/" {
						h.tunnel = u + "/"
						h.tunnelStat = "aktif"
						fmt.Println("Alamat INTERNET (bisa dibuka dari mana saja):")
						fmt.Println("   TV    :", h.tunnel+"#tv")
						fmt.Println("   Meja 1:", h.tunnel+"#meja1")
						fmt.Println("   (alamat ini berubah setiap kali program dijalankan ulang)")
						fmt.Println()
					}
					h.mu.Unlock()
				}
			}
		}
		go scan(stderr)
		go scan(stdout)
		cmd.Wait()
		h.mu.Lock()
		h.tunnel = ""
		h.tunnelStat = "terputus, mencoba lagi..."
		h.mu.Unlock()
		if attempt > 50 {
			return
		}
		time.Sleep(10 * time.Second)
	}
}
