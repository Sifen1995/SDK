import os
import joblib
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from sklearn.metrics import (
    classification_report,
    confusion_matrix,
    roc_curve,
    auc,
    precision_recall_curve,
)

from src.data.loader import DataLoader
from src.features.build_features import FeatureExtractor


class ModelEvaluator:
    """
    Executes in-depth evaluation on the held-out Test set:
    - Confusion Matrix visualization
    - ROC-AUC and Precision-Recall Curves
    - Feature Importance ranking (TF-IDF + Domain heuristics)
    """

    def __init__(self, artifact_dir: str = "models/artifacts"):
        self.artifact_dir = artifact_dir
        self.model = joblib.load(os.path.join(artifact_dir, "scam_detector.joblib"))
        self.vectorizer = joblib.load(os.path.join(artifact_dir, "tfidf_vectorizer.joblib"))
        self.config = joblib.load(os.path.join(artifact_dir, "feature_config.joblib"))
        self.feature_extractor = FeatureExtractor()

    def _prepare_matrix(self, df: pd.DataFrame):
        """Converts raw dataframe to combined feature matrix."""
        from scipy.sparse import hstack

        cleaned_texts = [
            self.feature_extractor.cleaner.clean_text(t, preserve_urls=True)
            for t in df["text"]
        ]
        tfidf_matrix = self.vectorizer.transform(cleaned_texts)
        tabular_df = self.feature_extractor.transform_dataframe(df, text_column="text")
        
        return hstack([tfidf_matrix, tabular_df.values]).tocsr(), list(tabular_df.columns)

    def evaluate_test_set(self, test_df: pd.DataFrame, output_dir: str = "reports"):
        """Runs evaluation suite and exports evaluation charts."""
        os.makedirs(output_dir, exist_ok=True)
        
        print("\n🔍 Evaluating Model on 100% Real Test Data...")
        X_test, tabular_names = self._prepare_matrix(test_df)
        y_test = test_df["label"].values

        # Predictions
        y_pred = self.model.predict(X_test)
        y_prob = self.model.predict_proba(X_test)[:, 1]

        # 1. Classification Report
        print("\n==========================================")
        print("📋 TEST SET METRICS REPORT")
        print("==========================================")
        print(classification_report(y_test, y_pred, target_names=["Ham (Legit)", "Spam (Scam)"]))

        # 2. Confusion Matrix Plot
        cm = confusion_matrix(y_test, y_pred)
        plt.figure(figsize=(6, 5))
        sns.heatmap(cm, annot=True, fmt="d", cmap="Blues", 
                    xticklabels=["Legit", "Scam"], 
                    yticklabels=["Legit", "Scam"])
        plt.title("Confusion Matrix - Test Set")
        plt.ylabel("Actual")
        plt.xlabel("Predicted")
        plt.tight_layout()
        plt.savefig(os.path.join(output_dir, "confusion_matrix.png"))
        plt.close()

        # 3. ROC Curve Plot
        fpr, tpr, _ = roc_curve(y_test, y_prob)
        roc_auc = auc(fpr, tpr)

        plt.figure(figsize=(7, 5))
        plt.plot(fpr, tpr, color="darkorange", lw=2, label=f"ROC curve (AUC = {roc_auc:.4f})")
        plt.plot([0, 1], [0, 1], color="navy", lw=2, linestyle="--")
        plt.xlim([0.0, 1.0])
        plt.ylim([0.0, 1.05])
        plt.xlabel("False Positive Rate")
        plt.ylabel("True Positive Rate")
        plt.title("Receiver Operating Characteristic (ROC)")
        plt.legend(loc="lower right")
        plt.grid(True, alpha=0.3)
        plt.savefig(os.path.join(output_dir, "roc_curve.png"))
        plt.close()

        # 4. Top Feature Importances
        all_feature_names = list(self.vectorizer.get_feature_names_out()) + tabular_names
        importances = self.model.feature_importances_
        
        top_indices = np.argsort(importances)[-15:][::-1]
        top_features = [all_feature_names[i] for i in top_indices]
        top_scores = importances[top_indices]

        plt.figure(figsize=(10, 6))
        sns.barplot(x=top_scores, y=top_features, palette="viridis")
        plt.title("Top 15 Most Important Model Features")
        plt.xlabel("Feature Importance Score")
        plt.tight_layout()
        plt.savefig(os.path.join(output_dir, "feature_importance.png"))
        plt.close()

        print(f"✅ Metric charts & reports saved to directory: {output_dir}/")


def main():
    # Load raw datasets and rebuild the same uncontaminated Train/Val/Test split
    # used during training (deterministic via random_state=42).
    loader = DataLoader(data_dir="data")
    _, _, test_df = loader.prepare_full_pipeline_data(
        test_size=0.1,
        val_size=0.1,
        random_state=42,
    )

    evaluator = ModelEvaluator(artifact_dir="models/artifacts")
    evaluator.evaluate_test_set(test_df, output_dir="reports")


if __name__ == "__main__":
    main()