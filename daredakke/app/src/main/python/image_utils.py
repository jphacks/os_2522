"""
画像処理用Pythonモジュール
PILとNumPyを使用した画像処理機能
"""

try:
    from PIL import Image
    import numpy as np
    import io
except ImportError as e:
    print(f"必要なライブラリがインストールされていません: {e}")

def convert_to_grayscale(image_bytes):
    """画像をグレースケールに変換"""
    try:
        # バイト配列から画像を読み込む
        image = Image.open(io.BytesIO(image_bytes))

        # グレースケールに変換
        grayscale_image = image.convert('L')

        # バイト配列に変換して返す
        output = io.BytesIO()
        grayscale_image.save(output, format='PNG')
        return output.getvalue()
    except Exception as e:
        raise RuntimeError(f"画像変換エラー: {e}")

def get_image_info(image_bytes):
    """画像の情報を取得"""
    try:
        image = Image.open(io.BytesIO(image_bytes))
        return {
            "width": image.width,
            "height": image.height,
            "format": image.format,
            "mode": image.mode
        }
    except Exception as e:
        raise RuntimeError(f"画像情報取得エラー: {e}")

def resize_image(image_bytes, width, height):
    """画像をリサイズ"""
    try:
        image = Image.open(io.BytesIO(image_bytes))
        resized_image = image.resize((width, height), Image.LANCZOS)

        output = io.BytesIO()
        resized_image.save(output, format='PNG')
        return output.getvalue()
    except Exception as e:
        raise RuntimeError(f"画像リサイズエラー: {e}")

def apply_blur(image_bytes, radius=2):
    """画像にぼかしを適用"""
    try:
        from PIL import ImageFilter

        image = Image.open(io.BytesIO(image_bytes))
        blurred_image = image.filter(ImageFilter.GaussianBlur(radius))

        output = io.BytesIO()
        blurred_image.save(output, format='PNG')
        return output.getvalue()
    except Exception as e:
        raise RuntimeError(f"ぼかし適用エラー: {e}")
