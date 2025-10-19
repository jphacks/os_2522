package com.example.daredakke.python

import android.content.Context
import com.chaquo.python.PyObject
import com.chaquo.python.Python
import com.chaquo.python.android.AndroidPlatform

/**
 * Chaquo Pythonを使用してKotlinからPythonコードを呼び出すヘルパークラス
 */
object PythonHelper {

    /**
     * Pythonインタープリタを初期化
     * アプリケーション起動時に一度だけ呼び出す必要がある
     */
    fun initialize(context: Context) {
        if (!Python.isStarted()) {
            Python.start(AndroidPlatform(context))
        }
    }

    /**
     * Pythonモジュールを取得
     */
    private fun getPython(): Python {
        return Python.getInstance()
    }

    /**
     * 計算機能のサンプル
     */
    object Calculator {
        private val module: PyObject by lazy {
            getPython().getModule("calculator")!!
        }

        fun add(a: Int, b: Int): Int {
            return module.callAttr("add", a, b).toInt()
        }

        fun subtract(a: Int, b: Int): Int {
            return module.callAttr("subtract", a, b).toInt()
        }

        fun multiply(a: Int, b: Int): Int {
            return module.callAttr("multiply", a, b).toInt()
        }

        fun divide(a: Double, b: Double): Double {
            return module.callAttr("divide", a, b)!!.toDouble()
        }

        fun getGreeting(name: String): String {
            return module.callAttr("get_greeting", name)!!.toString()
        }

        fun processList(numbers: List<Int>): Map<String, Double> {
            val result = module.callAttr("process_list", numbers)!!
            return mapOf(
                "sum" to result.get("sum")!!.toDouble(),
                "average" to result.get("average")!!.toDouble(),
                "max" to result.get("max")!!.toDouble(),
                "min" to result.get("min")!!.toDouble()
            )
        }
    }

    /**
     * 画像処理機能のサンプル
     */
    object ImageUtils {
        private val module: PyObject by lazy {
            getPython().getModule("image_utils")!!
        }

        fun convertToGrayscale(imageBytes: ByteArray): ByteArray {
            val result = module.callAttr("convert_to_grayscale", imageBytes)!!
            return result.toJava(ByteArray::class.java)
        }

        fun getImageInfo(imageBytes: ByteArray): Map<String, Any> {
            val result = module.callAttr("get_image_info", imageBytes)!!
            return mapOf(
                "width" to result.get("width")!!.toInt(),
                "height" to result.get("height")!!.toInt(),
                "format" to result.get("format")!!.toString(),
                "mode" to result.get("mode")!!.toString()
            )
        }

        fun resizeImage(imageBytes: ByteArray, width: Int, height: Int): ByteArray {
            val result = module.callAttr("resize_image", imageBytes, width, height)!!
            return result.toJava(ByteArray::class.java)
        }

        fun applyBlur(imageBytes: ByteArray, radius: Int = 2): ByteArray {
            val result = module.callAttr("apply_blur", imageBytes, radius)!!
            return result.toJava(ByteArray::class.java)
        }
    }

    /**
     * カスタムPythonコードを実行
     */
    fun executeCode(code: String): String {
        return try {
            val result = getPython().getModule("builtins")!!.callAttr("eval", code)
            result!!.toString()
        } catch (e: Exception) {
            "エラー: ${e.message}"
        }
    }

    /**
     * Pythonのバージョン情報を取得
     */
    fun getPythonVersion(): String {
        val sys = getPython().getModule("sys")!!
        return sys.get("version")!!.toString()
    }
}
