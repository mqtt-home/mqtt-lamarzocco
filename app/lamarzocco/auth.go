package lamarzocco

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

// InstallationKey holds the cryptographic material for API authentication
type InstallationKey struct {
	InstallationID string
	Secret         []byte // 32 bytes
	PrivateKey     *ecdsa.PrivateKey
}

// b64 encodes bytes to base64 string (standard encoding)
func b64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// GenerateInstallationKey creates a new installation key for API authentication
func GenerateInstallationKey() (*InstallationKey, error) {
	// Generate ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Generate installation ID (UUID)
	installationID := uuid.New().String()

	// Get public key in DER format
	pubKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Derive secret: SHA256(installation_id.pub_b64.inst_hash_b64)
	pubB64 := b64(pubKeyDER)
	instHash := sha256.Sum256([]byte(installationID))
	instHashB64 := b64(instHash[:])
	triple := fmt.Sprintf("%s.%s.%s", installationID, pubB64, instHashB64)
	secret := sha256.Sum256([]byte(triple))

	return &InstallationKey{
		InstallationID: installationID,
		Secret:         secret[:],
		PrivateKey:     privateKey,
	}, nil
}

// PublicKeyB64 returns the public key in base64-encoded DER format
func (k *InstallationKey) PublicKeyB64() (string, error) {
	pubKeyDER, err := x509.MarshalPKIXPublicKey(&k.PrivateKey.PublicKey)
	if err != nil {
		return "", err
	}
	return b64(pubKeyDER), nil
}

// BaseString returns the base string for proof generation
// Format: installation_id.base64(sha256(public_key_der_bytes))
func (k *InstallationKey) BaseString() string {
	pubKeyDER, _ := x509.MarshalPKIXPublicKey(&k.PrivateKey.PublicKey)
	pubKeyHash := sha256.Sum256(pubKeyDER)
	return k.InstallationID + "." + b64(pubKeyHash[:])
}

// GenerateRequestProof generates the request proof using La Marzocco's custom algorithm
func GenerateRequestProof(baseString string, secret []byte) string {
	if len(secret) != 32 {
		return ""
	}

	// Create mutable copy of 32-byte secret
	work := make([]byte, 32)
	copy(work, secret)

	baseBytes := []byte(baseString)

	for _, byteVal := range baseBytes {
		idx := int(byteVal) % 32
		shiftIdx := (idx + 1) % 32
		shiftAmount := work[shiftIdx] & 7 // 0-7 bit shift

		// XOR then rotate left
		xorResult := byteVal ^ work[idx]
		rotated := ((xorResult << shiftAmount) | (xorResult >> (8 - shiftAmount))) & 0xFF
		work[idx] = rotated
	}

	// Return base64(SHA256(final_work_array))
	hash := sha256.Sum256(work)
	return b64(hash[:])
}

// GenerateExtraHeaders generates the authentication headers for API requests
func (k *InstallationKey) GenerateExtraHeaders() (map[string]string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	nonce := strings.ToLower(uuid.New().String())

	// Generate proof using installation_id.nonce.timestamp as input
	proofInput := fmt.Sprintf("%s.%s.%s", k.InstallationID, nonce, timestamp)
	proof := GenerateRequestProof(proofInput, k.Secret)

	// Signature data: "{installation_id}.{nonce}.{timestamp}.{proof}"
	signatureData := fmt.Sprintf("%s.%s", proofInput, proof)

	// Sign with ECDSA-SHA256
	hash := sha256.Sum256([]byte(signatureData))
	r, s, err := ecdsa.Sign(rand.Reader, k.PrivateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	// Encode signature in DER format (ASN.1 SEQUENCE of two INTEGERs)
	signature := encodeDERSignature(r, s)
	signatureB64 := b64(signature)

	return map[string]string{
		"X-App-Installation-Id": k.InstallationID,
		"X-Timestamp":           timestamp,
		"X-Nonce":               nonce,
		"X-Request-Signature":   signatureB64,
	}, nil
}

// encodeDERSignature encodes ECDSA signature (r, s) in DER format (ASN.1)
func encodeDERSignature(r, s *big.Int) []byte {
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// Add leading zero if high bit is set (to ensure positive integer)
	if len(rBytes) > 0 && rBytes[0]&0x80 != 0 {
		rBytes = append([]byte{0x00}, rBytes...)
	}
	if len(sBytes) > 0 && sBytes[0]&0x80 != 0 {
		sBytes = append([]byte{0x00}, sBytes...)
	}

	// ASN.1 DER encoding: SEQUENCE { INTEGER r, INTEGER s }
	// 0x02 = INTEGER tag
	rPart := append([]byte{0x02, byte(len(rBytes))}, rBytes...)
	sPart := append([]byte{0x02, byte(len(sBytes))}, sBytes...)

	// 0x30 = SEQUENCE tag
	content := append(rPart, sPart...)
	return append([]byte{0x30, byte(len(content))}, content...)
}
