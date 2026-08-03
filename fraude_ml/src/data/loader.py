import os
import pandas as pd
import numpy as np
from typing import Tuple, Dict
from sklearn.model_selection import train_test_split


class DataLoader:
    """Handles loading, standardizing, balancing, and splitting of all raw datasets."""

    def __init__(self, data_dir: str):
        self.raw_dir = os.path.join(data_dir, "raw")
        self.processed_dir = os.path.join(data_dir, "processed")

    def load_uci_data(self, filename: str = "uci_spam.csv") -> pd.DataFrame:
        """Loads and standardizes UCI SMS Spam dataset."""
        filepath = os.path.join(self.raw_dir, filename)
        if not os.path.exists(filepath):
            print(f"⚠️ Warning: {filename} not found. Skipping.")
            return pd.DataFrame()

        # Handle tab-separated or comma-separated formats
        try:
            df = pd.read_csv(filepath, sep="\t", names=["raw_label", "text"], encoding="latin-1")
            if df["text"].isnull().all():
                df = pd.read_csv(filepath, encoding="latin-1")
        except Exception:
            df = pd.read_csv(filepath, encoding="latin-1")

        # Standardize labels (ham/0, spam/1)
        if "raw_label" in df.columns:
            df["label"] = df["raw_label"].astype(str).str.lower().map({"ham": 0, "spam": 1})
        elif "v1" in df.columns:
            df["label"] = df["v1"].astype(str).str.lower().map({"ham": 0, "spam": 1})
            df["text"] = df["v2"]

        df["source"] = "uci_public"
        df["is_synthetic"] = False
        return df[["text", "label", "source", "is_synthetic"]].dropna()

    def load_kaggle_data(self, filename: str = "kaggle_sms.csv") -> pd.DataFrame:
        """Loads and standardizes modern Kaggle SMS dataset."""
        filepath = os.path.join(self.raw_dir, filename)
        if not os.path.exists(filepath):
            print(f"⚠️ Warning: {filename} not found. Skipping.")
            return pd.DataFrame()

        df = pd.read_csv(filepath)
        # Dynamic column detection
        text_col = next((c for c in df.columns if c.lower() in ["text", "sms", "message", "body", "v2"]), None)
        label_col = next((c for c in df.columns if c.lower() in ["label", "target", "class", "category", "v1"]), None)

        if not text_col or not label_col:
            raise ValueError(f"Could not map text/label columns in {filename}")

        df["text"] = df[text_col]
        # Map string labels if necessary
        if df[label_col].dtype == object:
            df["label"] = df[label_col].astype(str).str.lower().map({"ham": 0, "spam": 1, "legit": 0, "smish": 1})
        else:
            df["label"] = df[label_col].astype(int)

        df["source"] = "kaggle_public"
        df["is_synthetic"] = False
        return df[["text", "label", "source", "is_synthetic"]].dropna()

    def load_local_synthetic_data(self, filename: str = "local_synthetic.csv") -> pd.DataFrame:
        """Loads generated synthetic local dataset."""
        filepath = os.path.join(self.raw_dir, filename)
        if not os.path.exists(filepath):
            raise FileNotFoundError(f"CRITICAL: {filename} missing! Run generate_local_data.py first.")

        df = pd.read_csv(filepath)
        df["source"] = "local_synthetic"
        df["is_synthetic"] = True
        return df[["text", "label", "source", "is_synthetic"]].dropna()

    def prepare_full_pipeline_data(
        self, test_size: float = 0.1, val_size: float = 0.1, random_state: int = 42
    ) -> Tuple[pd.DataFrame, pd.DataFrame, pd.DataFrame]:
        """
        Combines all sources and produces uncontaminated Train/Val/Test splits.
        Test set contains ONLY real public data + explicit test samples (0% synthetic contamination).
        """
        uci_df = self.load_uci_data()
        kaggle_df = self.load_kaggle_data()
        synthetic_df = self.load_local_synthetic_data()

        # Merge real public sources
        real_df = pd.concat([uci_df, kaggle_df], ignore_index=True).drop_duplicates(subset=["text"])

        # Split real data into Train, Val, and Test
        real_train, real_temp = train_test_split(
            real_df, test_size=(test_size + val_size), random_state=random_state, stratify=real_df["label"]
        )
        relative_val_ratio = val_size / (test_size + val_size)
        real_val, real_test = train_test_split(
            real_temp, test_size=(1.0 - relative_val_ratio), random_state=random_state, stratify=real_temp["label"]
        )

        # Synthetic data is injected ONLY into the training pipeline
        final_train = pd.concat([real_train, synthetic_df], ignore_index=True)
        final_train = final_train.drop_duplicates(subset=["text"]).sample(frac=1.0, random_state=random_state).reset_index(drop=True)

        os.makedirs(self.processed_dir, exist_ok=True)
        final_train.to_csv(os.path.join(self.processed_dir, "train_dataset.csv"), index=False)
        real_val.to_csv(os.path.join(self.processed_dir, "val_dataset.csv"), index=False)
        real_test.to_csv(os.path.join(self.processed_dir, "test_dataset.csv"), index=False)

        print(f"✅ Data Preparation Complete:")
        print(f"   - Train Pool: {len(final_train)} rows (Real: {len(real_train)}, Synthetic: {len(synthetic_df)})")
        print(f"   - Val Pool:   {len(real_val)} rows (100% Real)")
        print(f"   - Test Pool:  {len(real_test)} rows (100% Real - Uncontaminated Benchmark)")

        return final_train, real_val, real_test