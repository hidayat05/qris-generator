package qris

import (
	"strings"
	"testing"
)

func TestGenerateAndParse(t *testing.T) {
	data := &QRISData{
		PayloadFormatIndicator: "01",
		PointOfInitiationMethod: "11",
		MerchantAccountInfo: MerchantAccountInfo{
			GUID:       "54",
			MerchantID: "9876543210",
		},
		MerchantCategoryCode: "5211",
		TransactionCurrency:  "360",
		CountryCode:          "ID",
		MerchantName:         "TOKO MAKMUR",
		MerchantCity:         "JAKARTA",
	}

	raw := GenerateQRIS(data)
	if raw == "" {
		t.Fatal("GenerateQRIS returned empty string")
	}

	if !VerifyCRC(raw) {
		t.Errorf("CRC verification failed for generated QRIS: %s", raw)
	}

	parsed, err := ParseQRIS(raw)
	if err != nil {
		t.Fatalf("ParseQRIS failed: %v", err)
	}

	if parsed.MerchantName != data.MerchantName {
		t.Errorf("MerchantName = %q, want %q", parsed.MerchantName, data.MerchantName)
	}
	if parsed.MerchantCity != data.MerchantCity {
		t.Errorf("MerchantCity = %q, want %q", parsed.MerchantCity, data.MerchantCity)
	}
	if parsed.MerchantAccountInfo.MerchantID != data.MerchantAccountInfo.MerchantID {
		t.Errorf("MerchantID = %q, want %q", parsed.MerchantAccountInfo.MerchantID, data.MerchantAccountInfo.MerchantID)
	}
	if parsed.MerchantCategoryCode != data.MerchantCategoryCode {
		t.Errorf("MCC = %q, want %q", parsed.MerchantCategoryCode, data.MerchantCategoryCode)
	}
	if parsed.TransactionCurrency != data.TransactionCurrency {
		t.Errorf("Currency = %q, want %q", parsed.TransactionCurrency, data.TransactionCurrency)
	}
}

func TestGenerateDynamicQR(t *testing.T) {
	data := &QRISData{
		PayloadFormatIndicator: "01",
		PointOfInitiationMethod: "12",
		MerchantAccountInfo: MerchantAccountInfo{
			GUID:       "54",
			MerchantID: "1234567890",
		},
		MerchantCategoryCode: "5812",
		TransactionCurrency:  "360",
		TransactionAmount:    "50000",
		CountryCode:          "ID",
		MerchantName:         "WARUNG MAKAN",
		MerchantCity:         "BANDUNG",
	}

	raw := GenerateQRIS(data)
	if raw == "" {
		t.Fatal("GenerateQRIS returned empty string")
	}

	if !VerifyCRC(raw) {
		t.Errorf("CRC verification failed for dynamic QRIS: %s", raw)
	}

	parsed, err := ParseQRIS(raw)
	if err != nil {
		t.Fatalf("ParseQRIS failed: %v", err)
	}

	if parsed.TransactionAmount != "50000" {
		t.Errorf("Amount = %q, want %q", parsed.TransactionAmount, "50000")
	}
	if parsed.PointOfInitiationMethod != "12" {
		t.Errorf("Initiation method = %q, want %q", parsed.PointOfInitiationMethod, "12")
	}
}

func TestGenerateWithAdditionalData(t *testing.T) {
	data := &QRISData{
		PayloadFormatIndicator: "01",
		PointOfInitiationMethod: "11",
		MerchantAccountInfo: MerchantAccountInfo{
			GUID:       "54",
			MerchantID: "1122334455",
		},
		MerchantCategoryCode: "5211",
		TransactionCurrency:  "360",
		CountryCode:          "ID",
		MerchantName:         "MINIMARKET",
		MerchantCity:         "SURABAYA",
		AdditionalData: AdditionalData{
			StoreLabel: "MINI-001",
		},
	}

	raw := GenerateQRIS(data)
	if !VerifyCRC(raw) {
		t.Error("CRC should be valid")
	}

	parsed, err := ParseQRIS(raw)
	if err != nil {
		t.Fatalf("ParseQRIS failed: %v", err)
	}
	if parsed.AdditionalData.StoreLabel != "MINI-001" {
		t.Errorf("StoreLabel = %q, want %q", parsed.AdditionalData.StoreLabel, "MINI-001")
	}
}

func TestDecode(t *testing.T) {
	input := "000201010211"
	tlvs, err := Decode(input)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(tlvs) != 2 {
		t.Fatalf("expected 2 TLVs, got %d", len(tlvs))
	}
	if tlvs[0].Tag != "00" || tlvs[0].Value != "01" {
		t.Errorf("first TLV: got tag=%s val=%s", tlvs[0].Tag, tlvs[0].Value)
	}
	if tlvs[1].Tag != "01" || tlvs[1].Value != "11" {
		t.Errorf("second TLV: got tag=%s val=%s", tlvs[1].Tag, tlvs[1].Value)
	}
}

func TestDecodeContainerTags(t *testing.T) {
	data := &QRISData{
		PayloadFormatIndicator: "01",
		PointOfInitiationMethod: "11",
		MerchantAccountInfo: MerchantAccountInfo{
			GUID:       "54",
			MerchantID: "12345",
		},
		MerchantCategoryCode: "5211",
		TransactionCurrency:  "360",
		CountryCode:          "ID",
		MerchantName:         "TOKO",
		MerchantCity:         "KOTA",
	}

	raw := GenerateQRIS(data)
	tlvs, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	ma := FindTLV(tlvs, "02")
	if ma == nil {
		t.Fatal("tag 02 not found")
	}
	if len(ma.Children) == 0 {
		t.Fatal("tag 02 should have children")
	}
}

func TestDecodeRaw(t *testing.T) {
	raw := "000201010211"
	tlvs, err := DecodeRaw(raw)
	if err != nil {
		t.Fatalf("DecodeRaw failed: %v", err)
	}
	if len(tlvs) != 2 {
		t.Fatalf("expected 2 TLVs, got %d", len(tlvs))
	}
	for _, tlv := range tlvs {
		if len(tlv.Children) > 0 {
			t.Errorf("DecodeRaw should not set Children, got tag %s with children", tlv.Tag)
		}
	}
}

func TestCRCVerification(t *testing.T) {
	data := &QRISData{
		PayloadFormatIndicator: "01",
		PointOfInitiationMethod: "11",
		MerchantAccountInfo: MerchantAccountInfo{
			GUID:       "54",
			MerchantID: "12345",
		},
		MerchantCategoryCode: "5211",
		TransactionCurrency:  "360",
		CountryCode:          "ID",
		MerchantName:         "TEST",
		MerchantCity:         "CITY",
	}

	raw := GenerateQRIS(data)
	if !VerifyCRC(raw) {
		t.Error("CRC should be valid for generated string")
	}

	corrupted := raw[:len(raw)-4] + "FFFF"
	if VerifyCRC(corrupted) {
		t.Error("CRC should be invalid for corrupted string")
	}
}

func TestCRCTagHasNoChildren(t *testing.T) {
	data := &QRISData{
		PayloadFormatIndicator: "01",
		PointOfInitiationMethod: "11",
		MerchantAccountInfo: MerchantAccountInfo{
			GUID:       "54",
			MerchantID: "99999",
		},
		MerchantCategoryCode: "5211",
		TransactionCurrency:  "360",
		CountryCode:          "ID",
		MerchantName:         "TOKO",
		MerchantCity:         "KOTA",
	}

	raw := GenerateQRIS(data)
	tlvs, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	crc := FindTLV(tlvs, "63")
	if crc == nil {
		t.Fatal("tag 63 (CRC) not found")
	}
	if len(crc.Children) > 0 {
		t.Error("CRC should NOT have children")
	}
}

func TestFormatTLVs(t *testing.T) {
	data := &QRISData{
		PayloadFormatIndicator: "01",
		PointOfInitiationMethod: "11",
		MerchantAccountInfo: MerchantAccountInfo{
			GUID:       "54",
			MerchantID: "12345",
		},
		MerchantCategoryCode: "5211",
		TransactionCurrency:  "360",
		CountryCode:          "ID",
		MerchantName:         "TOKO",
		MerchantCity:         "KOTA",
	}

	raw := GenerateQRIS(data)
	tlvs, err := Decode(raw)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	formatted := FormatTLVs(tlvs, "")
	if !strings.Contains(formatted, "Payload Format Indicator") {
		t.Error("formatted output should contain field names")
	}
	if !strings.Contains(formatted, "Merchant Name") {
		t.Error("formatted output should contain 'Merchant Name'")
	}
}

func TestGenerateAndParseWithTipPrompt(t *testing.T) {
	data := &QRISData{
		PayloadFormatIndicator: "01",
		PointOfInitiationMethod: "11",
		MerchantAccountInfo: MerchantAccountInfo{
			GUID:       "54",
			MerchantID: "12345",
		},
		MerchantCategoryCode: "5211",
		TransactionCurrency:  "360",
		CountryCode:          "ID",
		MerchantName:         "TOKO TIP PROMPT",
		MerchantCity:         "JAKARTA",
		TipIndicator:         "01",
	}

	raw := GenerateQRIS(data)
	if !VerifyCRC(raw) {
		t.Error("CRC verification failed")
	}

	parsed, err := ParseQRIS(raw)
	if err != nil {
		t.Fatalf("ParseQRIS failed: %v", err)
	}

	if parsed.TipIndicator != "01" {
		t.Errorf("TipIndicator = %q, want %q", parsed.TipIndicator, "01")
	}
	if parsed.ConvenienceFeeFixed != "" {
		t.Errorf("ConvenienceFeeFixed should be empty, got %q", parsed.ConvenienceFeeFixed)
	}
	if parsed.ConvenienceFeePercentage != "" {
		t.Errorf("ConvenienceFeePercentage should be empty, got %q", parsed.ConvenienceFeePercentage)
	}
}

func TestGenerateAndParseWithTipFixed(t *testing.T) {
	data := &QRISData{
		PayloadFormatIndicator: "01",
		PointOfInitiationMethod: "11",
		MerchantAccountInfo: MerchantAccountInfo{
			GUID:       "54",
			MerchantID: "12345",
		},
		MerchantCategoryCode: "5211",
		TransactionCurrency:  "360",
		CountryCode:          "ID",
		MerchantName:         "TOKO TIP FIXED",
		MerchantCity:         "JAKARTA",
		TipIndicator:         "02",
		ConvenienceFeeFixed:  "5000",
	}

	raw := GenerateQRIS(data)
	if !VerifyCRC(raw) {
		t.Error("CRC verification failed")
	}

	parsed, err := ParseQRIS(raw)
	if err != nil {
		t.Fatalf("ParseQRIS failed: %v", err)
	}

	if parsed.TipIndicator != "02" {
		t.Errorf("TipIndicator = %q, want %q", parsed.TipIndicator, "02")
	}
	if parsed.ConvenienceFeeFixed != "5000" {
		t.Errorf("ConvenienceFeeFixed = %q, want %q", parsed.ConvenienceFeeFixed, "5000")
	}
	if parsed.ConvenienceFeePercentage != "" {
		t.Errorf("ConvenienceFeePercentage should be empty, got %q", parsed.ConvenienceFeePercentage)
	}
}

func TestGenerateAndParseWithTipPercentage(t *testing.T) {
	data := &QRISData{
		PayloadFormatIndicator: "01",
		PointOfInitiationMethod: "11",
		MerchantAccountInfo: MerchantAccountInfo{
			GUID:       "54",
			MerchantID: "12345",
		},
		MerchantCategoryCode: "5211",
		TransactionCurrency:  "360",
		CountryCode:          "ID",
		MerchantName:         "TOKO TIP PERCENT",
		MerchantCity:         "JAKARTA",
		TipIndicator:         "03",
		ConvenienceFeePercentage: "02.50",
	}

	raw := GenerateQRIS(data)
	if !VerifyCRC(raw) {
		t.Error("CRC verification failed")
	}

	parsed, err := ParseQRIS(raw)
	if err != nil {
		t.Fatalf("ParseQRIS failed: %v", err)
	}

	if parsed.TipIndicator != "03" {
		t.Errorf("TipIndicator = %q, want %q", parsed.TipIndicator, "03")
	}
	if parsed.ConvenienceFeePercentage != "02.50" {
		t.Errorf("ConvenienceFeePercentage = %q, want %q", parsed.ConvenienceFeePercentage, "02.50")
	}
	if parsed.ConvenienceFeeFixed != "" {
		t.Errorf("ConvenienceFeeFixed should be empty, got %q", parsed.ConvenienceFeeFixed)
	}
}
