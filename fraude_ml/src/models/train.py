import os
import joblib
import numpy as np
import pandas as pd
from scipy.sparse import hstack
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import (
    classification_report,
    roc_auc_score,
    confusion_matrix,
    accuracy_score,
    f1_score,
)

try:
    import lightgbm as lgb

    HAS_LIGHTGBM = True
except ImportError:
    lgb = None
    HAS_LIGHTGBM = False

from src.data.loader import DataLoader
from src.features.build_features import FeatureExtractor


class ModelTrainer:
    """
    Handles end-to-end training pipeline:
    - Feature extraction (TF-IDF + Tabular URL/Text heuristics)
    - Hybrid Matrix concatenation
    - Model training & hyperparameter tuning
    - Comprehensive performance evaluation
    - Artifact serialization
    """

    def __init__(self, artifact_dir: str = "models/artifacts"):
        self.artifact_dir = artifact_dir
        os.makedirs(self.artifact_dir, exist_ok=True)

        self.feature_extractor = FeatureExtractor()
        self.vectorizer = TfidfVectorizer(
            max_features=3000,
            ngram_range=(1, 2),
            sublinear_tf=True,
            strip_accents="unicode",
        )
        self.model = None
        self.feature_names = []

    def prepare_feature_matrix(
        self, df: pd.DataFrame, is_training: bool = False
    ):
        """
        Converts raw text into a combined hybrid feature matrix:
        [ TF-IDF Text Features (3000) ] + [ Structural & URL Features (13) ]
        """
        # 1. Cleaned Text TF-IDF Vectors
        cleaned_texts = [
            self.feature_extractor.cleaner.clean_text(t, preserve_urls=True)
            for t in df["text"]
        ]

        if is_training:
            tfidf_matrix = self.vectorizer.fit_transform(cleaned_texts)
        else:
            tfidf_matrix = self.vectorizer.transform(cleaned_texts)

        # 2. Extract Structural & URL Features
        tabular_df = self.feature_extractor.transform_dataframe(df, text_column="text")
        tabular_matrix = tabular_df.values

        if is_training:
            # Save tabular feature names for inference reference
            self.feature_names = list(tabular_df.columns)

        # 3. Stack TF-IDF sparse matrix with tabular features
        hybrid_matrix = hstack([tfidf_matrix, tabular_matrix]).tocsr()
        return hybrid_matrix

    def train_lightgbm(self, X_train, y_train, X_val, y_val):
        """Trains LightGBM if available; otherwise falls back to RandomForest."""
        if HAS_LIGHTGBM:
            print("\n🚀 Training LightGBM Classifier...")
            self.model = lgb.LGBMClassifier(
                n_estimators=500,
                learning_rate=0.03,
                max_depth=6,
                num_leaves=31,
                subsample=0.8,
                colsample_bytree=0.8,
                random_state=42,
                class_weight="balanced",
            )
            self.model.fit(
                X_train,
                y_train,
                eval_set=[(X_val, y_val)],
                callbacks=[lgb.early_stopping(stopping_rounds=30, verbose=False)],
            )
            return

        print("\n⚠️ lightgbm not installed — falling back to RandomForestClassifier...")
        self.model = RandomForestClassifier(
            n_estimators=200,
            max_depth=20,
            n_jobs=-1,
            random_state=42,
            class_weight="balanced",
        )
        self.model.fit(X_train, y_train)

    def evaluate(self, X, y, set_name: str = "Validation"):
        """Prints detailed performance evaluation report."""
        y_pred = self.model.predict(X)
        y_prob = self.model.predict_proba(X)[:, 1]

        acc = accuracy_score(y, y_pred)
        f1 = f1_score(y, y_pred)
        auc = roc_auc_score(y, y_prob)
        cm = confusion_matrix(y, y_pred)

        print(f"\n==========================================")
        print(f"📊 Performance Report: [{set_name.upper()} SET]")
        print(f"==========================================")
        print(f"Accuracy : {acc * 100:.2f}%")
        print(f"F1-Score : {f1:.4f}")
        print(f"ROC-AUC  : {auc:.4f}")
        print("\nClassification Report:")
        print(classification_report(y, y_pred, target_names=["Ham (0)", "Spam (1)"]))
        print("Confusion Matrix:")
        print(f"  [TN: {cm[0][0]}  FP: {cm[0][1]}]")
        print(f"  [FN: {cm[1][0]}  TP: {cm[1][1]}]")
        print(f"==========================================")

    def save_artifacts(self):
        """Saves vectorizer, feature configuration, and model weights."""
        model_path = os.path.join(self.artifact_dir, "scam_detector.joblib")
        vectorizer_path = os.path.join(self.artifact_dir, "tfidf_vectorizer.joblib")
        config_path = os.path.join(self.artifact_dir, "feature_config.joblib")

        joblib.dump(self.model, model_path)
        joblib.dump(self.vectorizer, vectorizer_path)
        joblib.dump({"tabular_features": self.feature_names}, config_path)

        print(f"\n✅ All artifacts successfully saved to: {self.artifact_dir}/")


def main():
    # 1. Load Data Splits (cwd must be fraude_ml/)
    loader = DataLoader(data_dir="data")
    train_df, val_df, test_df = loader.prepare_full_pipeline_data(
        test_size=0.1,
        val_size=0.1,
        random_state=42,
    )

    trainer = ModelTrainer(artifact_dir="models/artifacts")

    # 2. Build Feature Matrices
    print("\n⚙️ Building Hybrid Feature Matrices...")
    X_train = trainer.prepare_feature_matrix(train_df, is_training=True)
    X_val = trainer.prepare_feature_matrix(val_df, is_training=False)
    X_test = trainer.prepare_feature_matrix(test_df, is_training=False)

    y_train = train_df["label"].values
    y_val = val_df["label"].values
    y_test = test_df["label"].values

    print(f"Feature Matrix Shape: {X_train.shape} (Samples x Features)")

    # 3. Train Model
    trainer.train_lightgbm(X_train, y_train, X_val, y_val)

    # 4. Evaluate on Validation and Secret Test Sets
    trainer.evaluate(X_val, y_val, set_name="Validation")
    trainer.evaluate(X_test, y_test, set_name="Secret Test (100% Real)")

    # 5. Save Model Artifacts
    trainer.save_artifacts()


if __name__ == "__main__":
    main()
