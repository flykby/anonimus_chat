#!/bin/sh
set -e

cert="${WEBHOOK_CERT_PATH:-}"
key="${WEBHOOK_KEY_PATH:-}"

if [ -n "$cert" ] && [ -n "$key" ]; then
	if [ ! -r "$cert" ]; then
		echo "ERROR: certificate not readable: $cert" >&2
		echo "Fix on host: chmod 644 certs/webhook.pem" >&2
		exit 1
	fi
	if [ ! -r "$key" ]; then
		echo "ERROR: private key not readable by bot user (uid $(id -u)): $key" >&2
		echo "Fix on host: chmod 644 certs/webhook.key" >&2
		echo "Or regenerate: ./scripts/gen-webhook-cert.sh YOUR_PUBLIC_IP ./certs" >&2
		exit 1
	fi
fi

exec /usr/local/bin/bot
