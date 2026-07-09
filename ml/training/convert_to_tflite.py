# skykin-ml/training/convert_to_tflite.py

import numpy as np
import pandas as pd
import tensorflow as tf
import json
import os

def convert_to_tflite():
    print("Loading trained model...")
    model = tf.keras.models.load_model("../models/best_model.keras")

    # ── Conversion 1: Float32 (baseline, higher accuracy) ──────
    converter = tf.lite.TFLiteConverter.from_keras_model(model)
    tflite_float = converter.convert()

    float_path = "../models/intent_model_float32.tflite"
    with open(float_path, "wb") as f:
        f.write(tflite_float)
    float_size = os.path.getsize(float_path) / 1024
    print(f"Float32 model: {float_size:.1f} KB")

    # ── Conversion 2: INT8 Quantization (smaller, faster) ──────
    # Quantization reduces model from float32 to int8 weights
    # Typically 4x smaller with <1% accuracy loss

    # Need representative dataset for quantization calibration
    df = pd.read_csv("../data/processed/training_data.csv")
    feature_cols = [f"f_{i}" for i in range(47)]
    calibration_data = df[feature_cols].values[:500].astype(
        np.float32)

    def representative_dataset():
        for sample in calibration_data:
            yield [sample.reshape(1, -1)]

    converter_int8 = tf.lite.TFLiteConverter.from_keras_model(model)
    converter_int8.optimizations = [tf.lite.Optimize.DEFAULT]
    converter_int8.representative_dataset = representative_dataset
    converter_int8.target_spec.supported_ops = [
        tf.lite.OpsSet.TFLITE_BUILTINS_INT8
    ]
    converter_int8.inference_input_type  = tf.float32
    converter_int8.inference_output_type = tf.float32

    tflite_int8 = converter_int8.convert()

    int8_path = "../models/intent_model_int8.tflite"
    with open(int8_path, "wb") as f:
        f.write(tflite_int8)
    int8_size = os.path.getsize(int8_path) / 1024
    print(f"INT8 model:    {int8_size:.1f} KB")

    # ── Verify accuracy after quantization ─────────────────────
    print("\nVerifying quantized model accuracy...")

    interpreter = tf.lite.Interpreter(model_path=int8_path)
    interpreter.allocate_tensors()

    input_details  = interpreter.get_input_details()
    output_details = interpreter.get_output_details()

    test_samples = df[feature_cols].values[:1000].astype(np.float32)
    test_labels  = df["label_idx"].values[:1000]

    correct = 0
    for i, sample in enumerate(test_samples):
        interpreter.set_tensor(
            input_details[0]["index"],
            sample.reshape(1, -1)
        )
        interpreter.invoke()
        output = interpreter.get_tensor(
            output_details[0]["index"]
        )
        if np.argmax(output[0]) == test_labels[i]:
            correct += 1

    accuracy = correct / len(test_samples)
    print(f"Quantized accuracy: {accuracy:.4f}")

    if accuracy < 0.80:
        print("WARNING: Quantized accuracy below 80% threshold")
        print("Using float32 model instead")
        final_model = tflite_float
        final_path  = float_path
    else:
        print("Quantized model passes accuracy threshold")
        final_model = tflite_int8
        final_path  = int8_path

    # ── Copy to Flutter SDK assets ─────────────────────────────
    sdk_asset_path = (
        "../../skykin-sdk/lib/ml/assets/intent_model.tflite"
    )
    os.makedirs(os.path.dirname(sdk_asset_path), exist_ok=True)

    with open(sdk_asset_path, "wb") as f:
        f.write(final_model)

    final_size = os.path.getsize(sdk_asset_path) / 1024
    print(f"\nFinal model copied to SDK: {final_size:.1f} KB")

    # ── Copy label map to SDK ──────────────────────────────────
    sdk_label_path = (
        "../../skykin-sdk/lib/ml/assets/label_map.json"
    )
    with open("../models/label_map.json") as f:
        label_map = json.load(f)
    with open(sdk_label_path, "w") as f:
        json.dump(label_map, f)

    # ── Save model metadata ────────────────────────────────────
    metadata = {
        "model_version":    "1.0.0",
        "feature_size":     47,
        "num_classes":      len(label_map),
        "intent_classes":   list(label_map.values()),
        "quantized":        accuracy >= 0.80,
        "accuracy":         round(accuracy, 4),
        "model_size_kb":    round(final_size, 1),
        "confidence_threshold": 0.70,
    }

    metadata_path = (
        "../../skykin-sdk/lib/ml/assets/model_metadata.json"
    )
    with open(metadata_path, "w") as f:
        json.dump(metadata, f, indent=2)

    print(f"Metadata saved → {metadata_path}")
    print("\nConversion complete.")
    return metadata


if __name__ == "__main__":
    convert_to_tflite()