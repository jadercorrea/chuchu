#!/bin/bash
set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "🔧 Setting up Model Recommender..."
echo ""

if [ ! -d "venv" ]; then
    echo "📦 Creating virtual environment..."
    python3 -m venv venv
fi

echo "📥 Installing dependencies..."
./venv/bin/pip install -q --upgrade pip
./venv/bin/pip install -q -r requirements.txt

echo ""
echo "🎓 Training model..."
./venv/bin/python scripts/train.py

echo ""
echo "✅ Setup complete!"
echo ""
echo "Test the model:"
echo "  ./venv/bin/python scripts/predict.py"
echo ""
