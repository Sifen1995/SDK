import os
import joblib
import pandas as pd
from typing import List, Union
from sklearn.feature_extraction.text import TfidfVectorizer
from src.data.cleaner import TextCleaner


class SMSVectorizer:
    """
    TF-IDF Vectorizer capturing word and character subword n-grams.
    Saves and loads persistent model vocabularies.
    """

    def __init__(self, max_features: int = 3000):
        self.cleaner = TextCleaner()
        self.max_features = max_features
        # Word n-grams (1 to 2 words) + character n-grams (3 to 5 chars to catch typos)
        self.vectorizer = TfidfVectorizer(
            max_features=self.max_features,
            ngram_range=(1, 2),
            analyzer="word",
            sublinear_tf=True,
        )
        self.is_fitted = False

    def fit_transform(self, raw_texts: List[str]) -> pd.DataFrame:
        """Cleans input texts and fits the TF-IDF vocabulary matrix."""
        cleaned_texts = [self.cleaner.clean_text(t) for t in raw_texts]
        tfidf_matrix = self.vectorizer.fit_transform(cleaned_texts)
        self.is_fitted = True

        feature_names = [f"tfidf_{name}" for name in self.vectorizer.get_feature_names_out()]
        return pd.DataFrame(tfidf_matrix.toarray(), columns=feature_names)

    def transform(self, raw_texts: List[str]) -> pd.DataFrame:
        """Transforms new raw text samples using the pre-fitted vocabulary."""
        if not self.is_fitted:
            raise RuntimeError("Vectorizer must be fitted or loaded before calling transform().")

        cleaned_texts = [self.cleaner.clean_text(t) for t in raw_texts]
        tfidf_matrix = self.vectorizer.transform(cleaned_texts)

        feature_names = [f"tfidf_{name}" for name in self.vectorizer.get_feature_names_out()]
        return pd.DataFrame(tfidf_matrix.toarray(), columns=feature_names)

    def save(self, filepath: str):
        """Saves vectorizer instance to disk."""
        os.makedirs(os.path.dirname(filepath), exist_ok=True)
        joblib.dump(self.vectorizer, filepath)
        print(f"✅ Vectorizer saved to {filepath}")

    def load(self, filepath: str):
        """Loads vectorizer instance from disk."""
        if not os.path.exists(filepath):
            raise FileNotFoundError(f"Vectorizer file not found at {filepath}")
        self.vectorizer = joblib.load(filepath)
        self.is_fitted = True
        print(f"✅ Vectorizer loaded from {filepath}")