package id.bram.skortenismeja;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Context;
import android.content.Intent;
import android.net.wifi.WifiManager;
import android.os.Build;
import android.os.IBinder;
import android.os.PowerManager;
import java.io.File;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.net.Inet4Address;
import java.net.InetAddress;
import java.net.NetworkInterface;
import java.util.Enumeration;

public class ServerService extends Service {
    static final String CH = "server";
    static final String STOP = "id.bram.skortenismeja.STOP";
    volatile boolean running = false;
    volatile Process proc;
    Thread watcher;
    PowerManager.WakeLock wl;
    WifiManager.WifiLock wfl;

    @Override
    public IBinder onBind(Intent i) { return null; }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        if (intent != null && STOP.equals(intent.getAction())) {
            running = false;
            if (proc != null) proc.destroy();
            stopForeground(true);
            stopSelf();
            return START_NOT_STICKY;
        }
        startForeground(1, buildNote());
        if (wl == null) {
            PowerManager pm = (PowerManager) getSystemService(Context.POWER_SERVICE);
            wl = pm.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "skortenismeja:server");
            wl.acquire();
            WifiManager wm = (WifiManager) getApplicationContext().getSystemService(Context.WIFI_SERVICE);
            if (wm != null) {
                wfl = wm.createWifiLock(WifiManager.WIFI_MODE_FULL_HIGH_PERF, "skortenismeja:wifi");
                wfl.acquire();
            }
        }
        if (watcher == null) {
            running = true;
            watcher = new Thread(new Runnable() { public void run() { loop(); } });
            watcher.start();
        }
        return START_STICKY;
    }

    void loop() {
        File bin = new File(getApplicationInfo().nativeLibraryDir, "libturnamen.so");
        File data = new File(getFilesDir(), "data");
        data.mkdirs();
        File log = new File(getFilesDir(), "server.log");
        while (running) {
            try {
                ProcessBuilder pb = new ProcessBuilder(bin.getAbsolutePath(), "-port", "8080", "-nobrowser",
                    "-internet=false", "-data", data.getAbsolutePath(), "-ips", ipList());
                pb.redirectErrorStream(true);
                pb.directory(getFilesDir());
                proc = pb.start();
                InputStream in = proc.getInputStream();
                FileOutputStream out = new FileOutputStream(log, false);
                byte[] buf = new byte[4096];
                int n;
                while ((n = in.read(buf)) > 0) { out.write(buf, 0, n); out.flush(); }
                out.close();
                proc.waitFor();
            } catch (Exception e) {
                try {
                    FileOutputStream out = new FileOutputStream(log, true);
                    out.write(("\nERROR: " + e).getBytes());
                    out.close();
                } catch (Exception ignored) { }
            }
            if (!running) break;
            try { Thread.sleep(2000); } catch (InterruptedException e) { break; }
        }
    }

    static String ipList() {
        StringBuilder sb = new StringBuilder();
        try {
            Enumeration<NetworkInterface> en = NetworkInterface.getNetworkInterfaces();
            while (en != null && en.hasMoreElements()) {
                NetworkInterface ni = en.nextElement();
                if (!ni.isUp() || ni.isLoopback()) continue;
                Enumeration<InetAddress> as = ni.getInetAddresses();
                while (as.hasMoreElements()) {
                    InetAddress a = as.nextElement();
                    if (a instanceof Inet4Address && !a.isLoopbackAddress()) {
                        String ip = a.getHostAddress();
                        if (ip.startsWith("169.254.")) continue;
                        if (sb.length() > 0) sb.append(',');
                        sb.append(ip);
                    }
                }
            }
        } catch (Exception ignored) { }
        return sb.toString();
    }

    Notification buildNote() {
        NotificationManager nm = (NotificationManager) getSystemService(Context.NOTIFICATION_SERVICE);
        Notification.Builder b;
        if (Build.VERSION.SDK_INT >= 26) {
            NotificationChannel ch = new NotificationChannel(CH, "Server turnamen", NotificationManager.IMPORTANCE_LOW);
            nm.createNotificationChannel(ch);
            b = new Notification.Builder(this, CH);
        } else {
            b = new Notification.Builder(this);
        }
        Intent open = new Intent(this, MainActivity.class);
        open.setFlags(Intent.FLAG_ACTIVITY_NEW_TASK | Intent.FLAG_ACTIVITY_SINGLE_TOP);
        PendingIntent pOpen = PendingIntent.getActivity(this, 0, open, PendingIntent.FLAG_UPDATE_CURRENT);
        Intent stop = new Intent(this, ServerService.class);
        stop.setAction(STOP);
        PendingIntent pStop = PendingIntent.getService(this, 1, stop, PendingIntent.FLAG_UPDATE_CURRENT);
        String ips = ipList();
        String first = ips.length() > 0 ? ips.split(",")[0] : "127.0.0.1";
        b.setContentTitle("Server turnamen aktif")
         .setContentText("TV: http://" + first + ":8080/#tv")
         .setSmallIcon(android.R.drawable.stat_sys_upload_done)
         .setOngoing(true)
         .setContentIntent(pOpen)
         .addAction(android.R.drawable.ic_menu_close_clear_cancel, "Hentikan server", pStop);
        return b.build();
    }

    @Override
    public void onDestroy() {
        running = false;
        if (proc != null) proc.destroy();
        if (wl != null && wl.isHeld()) wl.release();
        if (wfl != null && wfl.isHeld()) wfl.release();
        super.onDestroy();
    }
}
