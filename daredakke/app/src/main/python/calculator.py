"""
サンプルPythonモジュール
Kotlinから呼び出される基本的な計算機能を提供
"""

def add(a, b):
    """2つの数値を足し算"""
    return a + b

def subtract(a, b):
    """2つの数値を引き算"""
    return a - b

def multiply(a, b):
    """2つの数値を掛け算"""
    return a * b

def divide(a, b):
    """2つの数値を割り算"""
    if b == 0:
        raise ValueError("ゼロで割ることはできません")
    return a / b

def get_greeting(name):
    """挨拶メッセージを返す"""
    return f"こんにちは、{name}さん！"

def process_list(numbers):
    """リストの合計、平均、最大、最小を計算"""
    if not numbers:
        return {"sum": 0, "average": 0, "max": 0, "min": 0}

    total = sum(numbers)
    return {
        "sum": total,
        "average": total / len(numbers),
        "max": max(numbers),
        "min": min(numbers)
    }
