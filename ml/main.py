"""Build the ML model and emit a final TensorFlow Lite artifact."""

from __future__ import annotations

import shutil
import subprocess
import sys
from pathlib import Path

ML_ROOT = Path(__file__).resolve().parent
TRAINING_DIR = ML_ROOT / "training"
MODELS_DIR = ML_ROOT / "models"
FINAL_TFLITE = MODELS_DIR / "final.tflite"


def run_step(script_name: str) -> None:
    script_path = TRAINING_DIR / script_name
    print(f"Running {script_path.name}...")
    subprocess.run([sys.executable, str(script_path)], cwd=TRAINING_DIR, check=True)


def build_model() -> Path:
    MODELS_DIR.mkdir(parents=True, exist_ok=True)

    run_step("generate_synthetic_data.py")
    run_step("train_model.py")
    run_step("convert_to_tflite.py")

    int8_path = MODELS_DIR / "intent_model_int8.tflite"
    float32_path = MODELS_DIR / "intent_model_float32.tflite"

    if int8_path.exists():
        shutil.copy2(int8_path, FINAL_TFLITE)
    elif float32_path.exists():
        shutil.copy2(float32_path, FINAL_TFLITE)
    else:
        raise FileNotFoundError("No TensorFlow Lite model was produced.")

    print(f"Final TensorFlow Lite model written to {FINAL_TFLITE}")
    return FINAL_TFLITE


if __name__ == "__main__":
    build_model()
