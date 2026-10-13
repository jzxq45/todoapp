package com.example.todoapp

import android.app.Activity
import android.net.Uri
import android.os.*
import android.view.*
import android.webkit.*
import android.widget.*
import com.example.todoapp.mobile.Mobile
import kotlinx.coroutines.*

class MainActivity : Activity() {
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Main)
    private lateinit var web: WebView
    private lateinit var progress: ProgressBar
    private lateinit var error: TextView
    private var port: Long? = null

    override fun onCreate(state: Bundle?) {
        super.onCreate(state)
        web = WebView(this).apply {
            settings.javaScriptEnabled = true
            settings.domStorageEnabled = true
            settings.allowFileAccess = false
            settings.allowContentAccess = false
            settings.mixedContentMode = android.webkit.WebSettings.MIXED_CONTENT_NEVER_ALLOW
            webViewClient = object : WebViewClient() {
                override fun shouldOverrideUrlLoading(view: WebView?, request: WebResourceRequest): Boolean {
                    val uri: Uri = request.url
                    return uri.scheme != "http" || uri.host != "127.0.0.1"
                }
                override fun onPageStarted(view: WebView?, url: String?, icon: android.graphics.Bitmap?) {
                    progress.setVisibility(View.VISIBLE)
                    error.visibility = View.GONE
                }
                override fun onPageFinished(view: WebView?, url: String?) { progress.setVisibility(View.GONE) }
                override fun onReceivedError(view: WebView?, request: WebResourceRequest, failure: android.webkit.WebResourceError) {
                    if (request.isForMainFrame) showError(getString(R.string.startup_error))
                }
            }
        }
        progress = ProgressBar(this, null, android.R.attr.progressBarStyleHorizontal).apply { isIndeterminate = true }
        error = TextView(this).apply {
            gravity = Gravity.CENTER
            setPadding(32, 24, 32, 24)
            visibility = View.GONE
            setOnClickListener { startServer() }
        }
        setContentView(FrameLayout(this).apply {
            addView(web, FrameLayout.LayoutParams(-1, -1))
            addView(progress, FrameLayout.LayoutParams(-1, 4, Gravity.TOP))
            addView(error, FrameLayout.LayoutParams(-1, -1))
        })
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            onBackInvokedDispatcher.registerOnBackInvokedCallback(
                android.window.OnBackInvokedDispatcher.PRIORITY_DEFAULT
            ) { goBack() }
        }
        startServer()
    }

    private fun startServer() {
        progress.setVisibility(View.VISIBLE)
        error.visibility = View.GONE
        scope.launch {
            try {
                val started = withContext(NonCancellable) {
                    val value = withContext(Dispatchers.IO) { Mobile.start(filesDir.absolutePath) }
                    if (isFinishing) withContext(Dispatchers.IO) { Mobile.stop() }
                    value
                }
                port = started
                if (!isFinishing) web.loadUrl("http://127.0.0.1:$started/")
            } catch (failure: Exception) {
                if (!isFinishing) showError("${getString(R.string.startup_error)}\n${failure.message.orEmpty()}")
            }
        }
    }

    private fun showError(message: String) {
        progress.setVisibility(View.GONE)
        error.text = "$message\n\n${getString(R.string.retry)}"
        error.visibility = View.VISIBLE
    }

    private fun goBack() {
        if (::web.isInitialized && web.canGoBack()) web.goBack() else finish()
    }

    @Deprecated("Use OnBackInvokedDispatcher on Android 13+")
    @Suppress("DEPRECATION")
    override fun onBackPressed() { goBack() }

    override fun onDestroy() {
        if (isFinishing && port != null) CoroutineScope(Dispatchers.IO).launch { Mobile.stop() }
        scope.cancel()
        if (::web.isInitialized) { web.stopLoading(); web.destroy() }
        super.onDestroy()
    }
}
