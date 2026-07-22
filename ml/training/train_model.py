# skykin-ml/training/train_model.py

import numpy as np
import pandas as pd
import tensorflow as tf
from tensorflow import keras
from sklearn.model_selection import train_test_split
from sklearn.metrics import classification_report, confusion_matrix
import json
import os

from feature_engineering import FEATURE_SIZE, INTENT_CLASSES

RANDOM_SEED   = 42
TEST_SIZE     = 0.20
BATCH_SIZE    = 256
EPOCHS        = 50
LEARNING_RATE = 0.001

tf.random.set_seed(RANDOM_SEED)
np.random.seed(RANDOM_SEED)

# Ensure output directory exists
os.makedirs("../models", exist_ok=True)

# ── Load data ──────────────────────────────────────────────────
print("Loading training data...")
df = pd.read_csv("../data/processed/training_data.csv")

feature_cols = [f"f_{i}" for i in range(FEATURE_SIZE)]

# Dynamic check to make sure feature size matches feature_engineering module
if len(feature_cols) != FEATURE_SIZE:
    raise ValueError(
        f"Data feature count mismatch! Expected {FEATURE_SIZE}, found {len(feature_cols)}."
    )

X = df[feature_cols].values.astype(np.float32)
y = df["label_idx"].values.astype(np.int32)

X_train, X_test, y_train, y_test = train_test_split(
    X, y,
    test_size=TEST_SIZE,
    random_state=RANDOM_SEED,
    stratify=y,
)

print(f"Train: {len(X_train):,} | Test: {len(X_test):,}")
print(f"Input Feature Dimensions: {X.shape[1]}")

# ── Build model ────────────────────────────────────────────────
# Architecture designed for Mobile/On-Device TFLite conversion:
# - Strictly static shapes matching (71,)
# - Standard dense layers + ReLU activations (broad platform delegate support)
# - Slightly widened Layer 1 (256 units) to process expanded 71-dim feature correlations
# - Target size after INT8/FP16 quantization remains under 1.5 MB

num_classes = len(INTENT_CLASSES)

model = keras.Sequential([
    keras.layers.Input(shape=(FEATURE_SIZE,), name="input"),

    # Layer 1 — Expanded from 128 to 256 to consume 71 cross-dimensional features
    keras.layers.Dense(256, name="dense_1"),
    keras.layers.BatchNormalization(name="bn_1"),
    keras.layers.Activation("relu", name="relu_1"),
    keras.layers.Dropout(0.3, name="dropout_1"),

    # Layer 2
    keras.layers.Dense(128, name="dense_2"),
    keras.layers.BatchNormalization(name="bn_2"),
    keras.layers.Activation("relu", name="relu_2"),
    keras.layers.Dropout(0.25, name="dropout_2"),

    # Layer 3
    keras.layers.Dense(64, name="dense_3"),
    keras.layers.BatchNormalization(name="bn_3"),
    keras.layers.Activation("relu", name="relu_3"),

    # Output Layer
    keras.layers.Dense(num_classes, activation="softmax", name="output"),
], name="skykin_intent_classifier")

model.summary()

# ── Compile ────────────────────────────────────────────────────
model.compile(
    optimizer=keras.optimizers.Adam(LEARNING_RATE),
    loss="sparse_categorical_crossentropy",
    metrics=["accuracy"],
)

# ── Callbacks ──────────────────────────────────────────────────
callbacks = [
    keras.callbacks.EarlyStopping(
        monitor="val_accuracy",
        patience=8,
        restore_best_weights=True,
        verbose=1,
    ),
    keras.callbacks.ReduceLROnPlateau(
        monitor="val_loss",
        factor=0.5,
        patience=4,
        min_lr=1e-5,
        verbose=1,
    ),
    keras.callbacks.ModelCheckpoint(
        "../models/best_model.keras",
        monitor="val_accuracy",
        save_best_only=True,
        verbose=1,
    ),
]

# ── Train ──────────────────────────────────────────────────────
print("\nTraining...")
history = model.fit(
    X_train, y_train,
    validation_split=0.15,
    epochs=EPOCHS,
    batch_size=BATCH_SIZE,
    callbacks=callbacks,
    verbose=1,
)

# ── Evaluate ───────────────────────────────────────────────────
print("\n── Test Set Evaluation ──────────────────────────────")
test_loss, test_acc = model.evaluate(X_test, y_test, verbose=0)
print(f"Accuracy: {test_acc:.4f}  Loss: {test_loss:.4f}")

y_pred = np.argmax(model.predict(X_test), axis=1)

print("\nClassification Report:")
print(classification_report(
    y_test, y_pred,
    target_names=INTENT_CLASSES,
))

print("\nConfusion Matrix:")
cm = confusion_matrix(y_test, y_pred)
cm_df = pd.DataFrame(
    cm,
    index=INTENT_CLASSES,
    columns=INTENT_CLASSES,
)
print(cm_df.to_string())

# Save label map for use in Flutter / On-Device SDK
label_map = {str(i): name for i, name in enumerate(INTENT_CLASSES)}
with open("../models/label_map.json", "w") as f:
    json.dump(label_map, f, indent=2)

print(f"\nLabel map saved → ../models/label_map.json")
print(f"Model saved    → ../models/best_model.keras")