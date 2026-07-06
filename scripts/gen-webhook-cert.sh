#!/bin/bash
set -e

# Generate self-signed certificate for Telegram webhook
# Usage: ./scripts/gen-webhook-cert.sh <IP_ADDRESS> [output_dir]
#
# Example:
#   ./scripts/gen-webhook-cert.sh 123.45.67.89
#   ./scripts/gen-webhook-cert.sh 123.45.67.89 /etc/ssl/webhook

IP_ADDRESS="${1:?Usage: $0 <IP_ADDRESS> [output_dir]}"
OUTPUT_DIR="${2:-./certs}"

mkdir -p "$OUTPUT_DIR"

CERT_PATH="$OUTPUT_DIR/webhook.pem"
KEY_PATH="$OUTPUT_DIR/webhook.key"

echo "Generating self-signed certificate for IP: $IP_ADDRESS"
echo "Output directory: $OUTPUT_DIR"

openssl req -newkey rsa:2048 -sha256 -nodes \
    -keyout "$KEY_PATH" \
    -x509 -days 365 \
    -out "$CERT_PATH" \
    -subj "/CN=$IP_ADDRESS" \
    2>/dev/null

# Bot container runs as uid 65534 and mounts ./certs read-only.
chmod 644 "$CERT_PATH" "$KEY_PATH"

echo ""
echo "Certificate generated successfully:"
echo "  Certificate: $CERT_PATH"
echo "  Private key: $KEY_PATH"
echo ""
echo "Add to .env:"
echo "  WEBHOOK_URL=https://$IP_ADDRESS:8443/telegram/webhook"
echo "  WEBHOOK_SECRET=$(openssl rand -hex 32)"
echo "  WEBHOOK_CERT_PATH=/app/certs/webhook.pem"
echo "  WEBHOOK_KEY_PATH=/app/certs/webhook.key"
echo ""
echo "Note: use your PUBLIC IP (curl -4 ifconfig.me), not the internal VM subnet."
echo "Note: Telegram supports ports 443, 80, 88, 8443 for webhooks"
