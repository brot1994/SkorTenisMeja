package id.bram.skortenismeja;

import android.app.Activity;
import android.content.Intent;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.view.KeyEvent;
import android.view.View;
import android.view.Window;
import android.view.WindowManager;
import android.webkit.WebResourceError;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.webkit.WebChromeClient;
import android.webkit.ValueCallback;
import android.webkit.DownloadListener;
import android.webkit.URLUtil;
import android.app.DownloadManager;
import android.content.Context;
import android.net.Uri;
import android.os.Environment;
import android.widget.Toast;
import java.net.InetSocketAddress;
import java.net.Socket;

public class MainActivity extends Activity {
    static final String BASE = "http://127.0.0.1:8080/";
    WebView web;
    final Handler ui = new Handler(Looper.getMainLooper());
    boolean loaded = false;
    ValueCallback<Uri[]> fileCb;

    @Override
    protected void onActivityResult(int req, int res, Intent data) {
        if (req == 7 && fileCb != null) {
            Uri[] r = null;
            if (res == RESULT_OK && data != null && data.getData() != null) r = new Uri[]{ data.getData() };
            fileCb.onReceiveValue(r);
            fileCb = null;
            return;
        }
        super.onActivityResult(req, res, data);
    }

    @Override
    protected void onCreate(Bundle b) {
        super.onCreate(b);
        requestWindowFeature(Window.FEATURE_NO_TITLE);
        getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);

        Intent svc = new Intent(this, ServerService.class);
        if (Build.VERSION.SDK_INT >= 26) startForegroundService(svc); else startService(svc);

        web = new WebView(this);
        WebSettings s = web.getSettings();
        s.setJavaScriptEnabled(true);
        s.setDomStorageEnabled(true);
        s.setDatabaseEnabled(true);
        s.setMediaPlaybackRequiresUserGesture(false);
        s.setUseWideViewPort(true);
        s.setLoadWithOverviewMode(true);
        web.setBackgroundColor(0xFF0F1318);
        web.setWebChromeClient(new WebChromeClient() {
            @Override
            public boolean onShowFileChooser(WebView v, ValueCallback<Uri[]> cb, FileChooserParams p) {
                if (fileCb != null) fileCb.onReceiveValue(null);
                fileCb = cb;
                Intent i = new Intent(Intent.ACTION_GET_CONTENT);
                i.addCategory(Intent.CATEGORY_OPENABLE);
                i.setType("*/*");
                try { startActivityForResult(Intent.createChooser(i, "Pilih file backup"), 7); }
                catch (Exception e) { fileCb = null; return false; }
                return true;
            }
        });
        web.setDownloadListener(new DownloadListener() {
            public void onDownloadStart(String url, String ua, String cd, String mime, long len) {
                try {
                    DownloadManager.Request r = new DownloadManager.Request(Uri.parse(url));
                    String name = URLUtil.guessFileName(url, cd, mime);
                    r.setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, name);
                    r.setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED);
                    ((DownloadManager) getSystemService(Context.DOWNLOAD_SERVICE)).enqueue(r);
                    Toast.makeText(MainActivity.this, "Backup disimpan ke folder Download: " + name, Toast.LENGTH_LONG).show();
                } catch (Exception e) {
                    Toast.makeText(MainActivity.this, "Gagal mengunduh: " + e.getMessage(), Toast.LENGTH_LONG).show();
                }
            }
        });
        web.setWebViewClient(new WebViewClient() {
            @Override
            public void onReceivedError(WebView v, WebResourceRequest req, WebResourceError err) {
                if (req != null && req.isForMainFrame()) retryLater();
            }
            @Override
            public void doUpdateVisitedHistory(WebView v, String url, boolean reload) {
                super.doUpdateVisitedHistory(v, url, reload);
                applyFullscreen(url != null && url.endsWith("#tv"));
            }
        });
        web.loadDataWithBaseURL(null, "<html><body style='background:#0f1318;color:#eef1f5;font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:90vh'><h2>Menyalakan server turnamen…</h2></body></html>", "text/html", "utf-8", null);
        setContentView(web);
        waitForServer();
    }

    void waitForServer() {
        new Thread(new Runnable() {
            public void run() {
                for (int i = 0; i < 120; i++) {
                    try {
                        Socket so = new Socket();
                        so.connect(new InetSocketAddress("127.0.0.1", 8080), 500);
                        so.close();
                        ui.post(new Runnable() { public void run() { if (!loaded) { loaded = true; web.loadUrl(BASE); } } });
                        return;
                    } catch (Exception e) {
                        try { Thread.sleep(500); } catch (InterruptedException ie) { return; }
                    }
                }
            }
        }).start();
    }

    void retryLater() {
        ui.postDelayed(new Runnable() { public void run() { web.reload(); } }, 1500);
    }

    void applyFullscreen(boolean on) {
        View d = getWindow().getDecorView();
        if (on) {
            d.setSystemUiVisibility(View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY | View.SYSTEM_UI_FLAG_FULLSCREEN
                | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION | View.SYSTEM_UI_FLAG_LAYOUT_STABLE
                | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN);
        } else {
            d.setSystemUiVisibility(View.SYSTEM_UI_FLAG_VISIBLE);
        }
    }

    @Override
    public boolean onKeyDown(int keyCode, KeyEvent event) {
        if (keyCode == KeyEvent.KEYCODE_BACK) {
            if (web != null && web.canGoBack()) { web.goBack(); return true; }
            moveTaskToBack(true);
            return true;
        }
        return super.onKeyDown(keyCode, event);
    }

    @Override
    protected void onDestroy() {
        if (web != null) web.destroy();
        super.onDestroy();
    }
}
