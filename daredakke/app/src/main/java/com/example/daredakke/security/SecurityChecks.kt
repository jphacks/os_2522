package com.example.daredakke.security

import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import java.io.BufferedReader
import java.io.File
import java.io.InputStreamReader
import java.net.InetSocketAddress
import java.net.Socket

object SecurityChecks {
    fun isRooted(context: Context): Boolean {
        return hasSuBinary() || canExecuteSu() || containsTestKeys() || hasDangerousPackages(context)
    }

    fun isFridaDetected(): Boolean {
        return tracerPidDetected() || mapsContainsFrida() || fridaServerPortOpen()
    }

    private fun hasSuBinary(): Boolean {
        val paths = arrayOf(
            "/system/bin/su","/system/xbin/su","/sbin/su",
            "/system/sd/xbin/su","/system/bin/failsafe/su",
            "/data/local/su","/data/local/xbin/su","/data/local/bin/su"
        )
        return paths.any { File(it).exists() }
    }

    private fun canExecuteSu(): Boolean = try {
        val p = Runtime.getRuntime().exec(arrayOf("/system/xbin/which", "su"))
        BufferedReader(InputStreamReader(p.inputStream)).use { it.readLine()?.isNotEmpty() == true }
    } catch (_: Throwable) { false }

    private fun containsTestKeys(): Boolean = Build.TAGS?.contains("test-keys") == true

    private fun hasDangerousPackages(context: Context): Boolean {
        val pm = context.packageManager
        val pkgs = listOf(
            "com.topjohnwu.magisk","eu.chainfire.supersu","com.noshufou.android.su",
            "com.koushikdutta.superuser","com.zachspong.temprootremovejb","com.ramdroid.appquarantine"
        )
        return pkgs.any {
            try { pm.getPackageInfo(it, 0); true } catch (_: PackageManager.NameNotFoundException) { false } catch (_: Throwable) { false }
        }
    }

 // ...existing code...
    private fun tracerPidDetected(): Boolean {
        return try {
            val status = File("/proc/self/status")
            if (!status.exists()) return false
            val tracerLine = status.readLines().firstOrNull { it.startsWith("TracerPid:") }
            val v = tracerLine?.substringAfter("TracerPid:")?.trim()?.toIntOrNull() ?: 0
            v > 0
        } catch (_: Throwable) {
            false
        }
    }

    private fun mapsContainsFrida(): Boolean {
        return try {
            val maps = File("/proc/self/maps")
            if (!maps.exists()) return false
            maps.useLines { seq ->
                seq.any { line ->
                    val s = line.lowercase()
                    s.contains("frida") || s.contains("gadget") || s.contains("gum-js")
                }
            }
        } catch (_: Throwable) {
            false
        }
    }
// ...existing code...

    private fun fridaServerPortOpen(): Boolean {
        val ports = listOf(27042, 27043)
        return ports.any { p ->
            try { Socket().use { it.connect(InetSocketAddress("127.0.0.1", p), 150); true } }
            catch (_: Throwable) { false }
        }
    }
}