import os
import random
import pandas as pd
from typing import List, Dict, Any


class LocalDataGenerator:
    """
    Generates realistic local Ethiopian SMS datasets (Amharic & English)
    combining legitimate service notifications with localized phishing/scam lures.
    """

    def __init__(self, phishtank_path: str = None):
        self.phishing_domains = self._load_phishing_domains(phishtank_path)

    def _load_phishing_domains(self, filepath: str) -> List[str]:
        """Loads domain structures from PhishTank if available, else falls back to generated ones."""
        if filepath and os.path.exists(filepath):
            try:
                df = pd.read_csv(filepath)
                if "url" in df.columns:
                    urls = df["url"].dropna().sample(min(200, len(df))).tolist()
                    print(f"✅ Loaded {len(urls)} phishing domain seeds from PhishTank.")
                    return urls
            except Exception as e:
                print(f"⚠️ Could not load PhishTank CSV ({e}). Using generated phishing seeds.")

        # Fallback realistic scam domains
        return [
            "http://telebirr-verify-kyc.xyz/login.php",
            "http://cbe-birr-unblock.top/account",
            "http://ethiotel-bonus-5g.sbs/claim",
            "http://192.168.1.105/cbe-sec",
            "http://telebirr-reward-2026.cfd/auth",
            "http://cbe-online-fix.info/verify",
            "http://bit.ly/ethio-bonus-claim",
        ]

    def _get_random_phish_url(self) -> str:
        return random.choice(self.phishing_domains)

    def generate_ham_samples(self, count: int = 500) -> List[Dict[str, Any]]:
        """Generates legitimate (Ham = 0) transaction and utility notifications."""
        templates = [
            # Telebirr Ham
            "ከ Telebirr: 500.00 ETB ለ {name} (0911***{digit}) ተልኳል። Transaction ID: TX{tx_id}. አጠቃላይ ቀሪ ሒሳብዎ 2450.50 ETB ነው።",
            "Telebirr: You have successfully bought 2GB 7-Days Data Package. Transaction ID: TX{tx_id}. Thank you for using Ethio Telecom.",
            "ከ 127: ውድ ደንበኛችን የ 100 ETB የአየር ሰዓት ሞልተዋል። ቀሪ ሂሳብዎ 120.50 ETB ነው።",
            # CBE Ham
            "Dear Customer, ETB 1,500.00 credited to account 1000****{digit} from {name}. Ref No: FT{tx_id}. CBE",
            "Dear Customer, ETB 200.00 debited from account 1000****{digit} for ATM Withdrawal. Ref No: FT{tx_id}. CBE",
            # Ethio Telecom / General Ham
            "Your Ethio Telecom package will expire in 2 days. Dial *999# to renew your active monthly internet bundle.",
            "Dear customer, your request for bill payment of 450 ETB has been completed. Thank you.",
        ]

        names = ["Abebe Kebede", "Tigist Alemu", "Dawit Solomon", "Marta Tadesse", "Sifen G."]
        data = []

        for _ in range(count):
            tpl = random.choice(templates)
            text = tpl.format(
                name=random.choice(names),
                digit=random.randint(100, 999),
                tx_id=random.randint(10000000, 99999999),
            )
            data.append({"text": text, "label": 0, "source": "local_synthetic"})
        return data

    def generate_spam_samples(self, count: int = 500) -> List[Dict[str, Any]]:
        """Generates localized phishing and scam (Spam = 1) messages."""
        templates = [
            # Telebirr Phishing
            "ማስጠንቀቂያ: የ Telebirr መለያዎ ታግዷል። እባክዎን በዚህ ሊንክ በመግባት መለያዎን ያረጋግጡ: {url}",
            "ከ Ethio Telecom: 1000 ETB እና 5GB ነፃ የኢንተርኔት ስጦታ አሸንፈዋል! አሁኑኑ ለመቀበል: {url}",
            "Urgent Telebirr Alert: Your wallet is suspended due to unverified KYC. Click here to update immediately: {url}",
            # CBE Phishing
            "CBE Urgent Alert: Your mobile banking account 1000****{digit} has been locked. Verify identity at: {url}",
            "Dear CBE Customer, suspicious login detected on your account. Restore access now: {url}",
            # Financial Scams
            "Congratulations! You were selected for a 50,000 ETB Ethio Telecom promo reward. Enter phone number to claim: {url}",
            "ከ 994: የእርስዎ መለያ ለደህንነት ሲባል ተዘግቷል። ለማስተካከል ይጫኑ: {url}",
        ]

        data = []
        for _ in range(count):
            tpl = random.choice(templates)
            text = tpl.format(
                digit=random.randint(100, 999),
                url=self._get_random_phish_url(),
            )
            data.append({"text": text, "label": 1, "source": "local_synthetic"})
        return data

    def generate_and_save(self, output_path: str, count_per_class: int = 500):
        """Generates combined local dataset and writes to disk."""
        os.makedirs(os.path.dirname(output_path), exist_ok=True)
        ham = self.generate_ham_samples(count_per_class)
        spam = self.generate_spam_samples(count_per_class)

        combined = ham + spam
        random.shuffle(combined)

        df = pd.DataFrame(combined)
        df.to_csv(output_path, index=False)
        print(f"✅ Generated {len(df)} local Ethiopian SMS samples at: {output_path}")


if __name__ == "__main__":
    generator = LocalDataGenerator(phishtank_path="../../data/raw/phishtank.csv")
    generator.generate_and_save("../../data/raw/local_synthetic.csv", count_per_class=500)