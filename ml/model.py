"""
Deprecated — use training pipeline instead.

  cd ml
  python3 -m training.train
"""

if __name__ == "__main__":
    import subprocess
    import sys

    subprocess.check_call([sys.executable, "-m", "training.train"])
