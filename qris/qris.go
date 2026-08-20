package qris

import (
	"fmt"
	"strings"
)

const (
	DefaultGUID             = "54"
	DefaultMCC              = "5211"
	DefaultCurrency         = "360"
	DefaultCountryCode      = "ID"
	StaticQR                = "11"
	DynamicQR               = "12"
	DefaultPayloadIndicator = "01"
)

type QRISData struct {
	PayloadFormatIndicator string
	PointOfInitiationMethod string
	MerchantAccountInfo     MerchantAccountInfo
	MerchantCategoryCode    string
	TransactionCurrency     string
	TransactionAmount       string
	CountryCode             string
	MerchantName            string
	MerchantCity            string
	PostalCode              string
	TipIndicator            string
	ConvenienceFeeFixed     string
	ConvenienceFeePercentage string
	AdditionalData          AdditionalData
}

type MerchantAccountInfo struct {
	GUID       string
	PAN        string
	MerchantID string
}

type AdditionalData struct {
	BillNumber    string
	MobileNumber  string
	StoreLabel    string
	LoyaltyNumber string
	TaxLabel      string
}

func ParseQRIS(input string) (*QRISData, error) {
	tlvs, err := Decode(input)
	if err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	data := &QRISData{}

	if t := FindTLV(tlvs, "00"); t != nil {
		data.PayloadFormatIndicator = t.Value
	}
	if t := FindTLV(tlvs, "01"); t != nil {
		data.PointOfInitiationMethod = t.Value
	}

	for _, tag := range []string{"02", "03", "50", "51"} {
		if t := FindTLV(tlvs, tag); t != nil && len(t.Children) > 0 {
			if c := FindTLV(t.Children, "00"); c != nil {
				data.MerchantAccountInfo.GUID = c.Value
			}
			if c := FindTLV(t.Children, "01"); c != nil {
				data.MerchantAccountInfo.PAN = c.Value
			}
			if c := FindTLV(t.Children, "02"); c != nil {
				data.MerchantAccountInfo.MerchantID = c.Value
			}
			break
		}
	}

	if t := FindTLV(tlvs, "52"); t != nil {
		data.MerchantCategoryCode = t.Value
	}
	if t := FindTLV(tlvs, "53"); t != nil {
		data.TransactionCurrency = t.Value
	}
	if t := FindTLV(tlvs, "54"); t != nil {
		data.TransactionAmount = t.Value
	}
	if t := FindTLV(tlvs, "55"); t != nil {
		data.TipIndicator = t.Value
	}
	if t := FindTLV(tlvs, "56"); t != nil {
		data.ConvenienceFeeFixed = t.Value
	}
	if t := FindTLV(tlvs, "57"); t != nil {
		data.ConvenienceFeePercentage = t.Value
	}
	if t := FindTLV(tlvs, "58"); t != nil {
		data.CountryCode = t.Value
	}
	if t := FindTLV(tlvs, "59"); t != nil {
		data.MerchantName = t.Value
	}
	if t := FindTLV(tlvs, "60"); t != nil {
		data.MerchantCity = t.Value
	}
	if t := FindTLV(tlvs, "61"); t != nil {
		data.PostalCode = t.Value
	}

	if t := FindTLV(tlvs, "62"); t != nil {
		if c := FindTLV(t.Children, "01"); c != nil {
			data.AdditionalData.BillNumber = c.Value
		}
		if c := FindTLV(t.Children, "02"); c != nil {
			data.AdditionalData.MobileNumber = c.Value
		}
		if c := FindTLV(t.Children, "03"); c != nil {
			data.AdditionalData.StoreLabel = c.Value
		}
		if c := FindTLV(t.Children, "04"); c != nil {
			data.AdditionalData.LoyaltyNumber = c.Value
		}
		if c := FindTLV(t.Children, "05"); c != nil {
			data.AdditionalData.TaxLabel = c.Value
		}
	}

	return data, nil
}

func GenerateQRIS(data *QRISData) string {
	var tlvs []TLV

	if data.PayloadFormatIndicator == "" {
		data.PayloadFormatIndicator = DefaultPayloadIndicator
	}
	tlvs = append(tlvs, TLV{Tag: "00", Value: data.PayloadFormatIndicator})

	if data.PointOfInitiationMethod == "" {
		data.PointOfInitiationMethod = StaticQR
	}
	tlvs = append(tlvs, TLV{Tag: "01", Value: data.PointOfInitiationMethod})

	var merchantChildren []TLV
	if data.MerchantAccountInfo.GUID != "" {
		merchantChildren = append(merchantChildren, TLV{Tag: "00", Value: data.MerchantAccountInfo.GUID})
	}
	if data.MerchantAccountInfo.PAN != "" {
		merchantChildren = append(merchantChildren, TLV{Tag: "01", Value: data.MerchantAccountInfo.PAN})
	}
	if data.MerchantAccountInfo.MerchantID != "" {
		merchantChildren = append(merchantChildren, TLV{Tag: "02", Value: data.MerchantAccountInfo.MerchantID})
	}
	if len(merchantChildren) > 0 {
		var value string
		for _, c := range merchantChildren {
			value += c.Encoded()
		}
		tlvs = append(tlvs, TLV{Tag: "02", Value: value, Children: merchantChildren})
	}

	if data.MerchantCategoryCode == "" {
		data.MerchantCategoryCode = DefaultMCC
	}
	tlvs = append(tlvs, TLV{Tag: "52", Value: data.MerchantCategoryCode})

	if data.TransactionCurrency == "" {
		data.TransactionCurrency = DefaultCurrency
	}
	tlvs = append(tlvs, TLV{Tag: "53", Value: data.TransactionCurrency})

	if data.TransactionAmount != "" {
		tlvs = append(tlvs, TLV{Tag: "54", Value: data.TransactionAmount})
	}

	if data.TipIndicator != "" {
		tlvs = append(tlvs, TLV{Tag: "55", Value: data.TipIndicator})
		if data.TipIndicator == "02" && data.ConvenienceFeeFixed != "" {
			tlvs = append(tlvs, TLV{Tag: "56", Value: data.ConvenienceFeeFixed})
		} else if data.TipIndicator == "03" && data.ConvenienceFeePercentage != "" {
			tlvs = append(tlvs, TLV{Tag: "57", Value: data.ConvenienceFeePercentage})
		}
	}

	if data.CountryCode == "" {
		data.CountryCode = DefaultCountryCode
	}
	tlvs = append(tlvs, TLV{Tag: "58", Value: data.CountryCode})
	tlvs = append(tlvs, TLV{Tag: "59", Value: data.MerchantName})
	tlvs = append(tlvs, TLV{Tag: "60", Value: data.MerchantCity})

	if data.PostalCode != "" {
		tlvs = append(tlvs, TLV{Tag: "61", Value: data.PostalCode})
	}

	var additionalChildren []TLV
	if data.AdditionalData.BillNumber != "" {
		additionalChildren = append(additionalChildren, TLV{Tag: "01", Value: data.AdditionalData.BillNumber})
	}
	if data.AdditionalData.MobileNumber != "" {
		additionalChildren = append(additionalChildren, TLV{Tag: "02", Value: data.AdditionalData.MobileNumber})
	}
	if data.AdditionalData.StoreLabel != "" {
		additionalChildren = append(additionalChildren, TLV{Tag: "03", Value: data.AdditionalData.StoreLabel})
	}
	if data.AdditionalData.LoyaltyNumber != "" {
		additionalChildren = append(additionalChildren, TLV{Tag: "04", Value: data.AdditionalData.LoyaltyNumber})
	}
	if data.AdditionalData.TaxLabel != "" {
		additionalChildren = append(additionalChildren, TLV{Tag: "05", Value: data.AdditionalData.TaxLabel})
	}
	if len(additionalChildren) > 0 {
		var value string
		for _, c := range additionalChildren {
			value += c.Encoded()
		}
		tlvs = append(tlvs, TLV{Tag: "62", Value: value, Children: additionalChildren})
	}

	var result strings.Builder
	for _, tlv := range tlvs {
		result.WriteString(tlv.Encoded())
	}

	crcInput := result.String() + "6304"
	crc := calculateCRC(crcInput)
	result.WriteString("6304" + crc)

	return result.String()
}

func calculateCRC(data string) string {
	crc := uint16(0xFFFF)
	for i := 0; i < len(data); i++ {
		crc ^= uint16(data[i]) << 8
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	crc &= 0xFFFF
	return fmt.Sprintf("%04X", crc)
}

func VerifyCRC(input string) bool {
	if len(input) < 8 {
		return false
	}
	crcStr := input[len(input)-4:]
	data := input[:len(input)-4]
	expectedCRC := calculateCRC(data)
	return crcStr == expectedCRC
}
