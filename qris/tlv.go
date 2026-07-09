package qris

import (
	"fmt"
	"strconv"
	"strings"
)

type TLV struct {
	Tag      string
	Value    string
	Children []TLV
}

var tagNames = map[string]string{
	"00": "Payload Format Indicator",
	"01": "Point of Initiation Method",
	"02": "Merchant Account Information",
	"03": "Merchant Account Information",
	"04": "Merchant Account Information",
	"05": "Merchant Account Information",
	"06": "Merchant Account Information",
	"07": "Merchant Account Information",
	"08": "Merchant Account Information",
	"09": "Additional Data",
	"50": "Merchant Account Information",
	"51": "Merchant Account Information",
	"52": "Merchant Category Code",
	"53": "Transaction Currency",
	"54": "Transaction Amount",
	"55": "Tip or Convenience Indicator",
	"56": "Convenience Fee Percentage",
	"57": "Convenience Fee Fixed",
	"58": "Country Code",
	"59": "Merchant Name",
	"60": "Merchant City",
	"61": "Postal Code",
	"62": "Additional Data",
	"63": "CRC",
	"64": "Merchant Information Language",
}

var subTagNames = map[string]string{
	"00": "GUI (Globally Unique Identifier)",
	"01": "PAN / Merchant Account",
	"02": "Merchant ID",
}

var additionalDataNames = map[string]string{
	"01": "Bill Number",
	"02": "Mobile Number",
	"03": "Store Label",
	"04": "Loyalty Number",
	"05": "Reference Label",
	"06": "Customer Label",
	"07": "Terminal Label",
	"08": "Purpose of Transaction",
	"09": "Additional Consumer Data",
}

var containerTags = map[string]bool{
	"02": true, "03": true, "04": true, "05": true,
	"06": true, "07": true, "08": true,
	"50": true, "51": true,
	"62": true,
	"09": true,
	"64": true,
}

func tagName(tag string) string {
	if name, ok := tagNames[tag]; ok {
		return fmt.Sprintf("%s (%s)", tag, name)
	}
	return tag
}

func subTagName(tag, parentTag string) string {
	if parentTag == "62" || parentTag == "09" {
		if name, ok := additionalDataNames[tag]; ok {
			return fmt.Sprintf("%s (%s)", tag, name)
		}
	}
	if name, ok := subTagNames[tag]; ok {
		return fmt.Sprintf("%s (%s)", tag, name)
	}
	return tag
}

func (t TLV) Encoded() string {
	if len(t.Children) > 0 {
		var value strings.Builder
		for _, child := range t.Children {
			value.WriteString(child.Encoded())
		}
		return fmt.Sprintf("%s%02d%s", t.Tag, value.Len(), value.String())
	}
	return fmt.Sprintf("%s%02d%s", t.Tag, len(t.Value), t.Value)
}

func Decode(input string) ([]TLV, error) {
	var tlvs []TLV
	for i := 0; i < len(input); {
		if i+4 > len(input) {
			return nil, fmt.Errorf("unexpected end at position %d", i)
		}
		tag := input[i : i+2]
		lengthStr := input[i+2 : i+4]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid length '%s' at position %d", lengthStr, i+2)
		}

		if i+4+length > len(input) {
			return nil, fmt.Errorf("value length %d exceeds remaining input at position %d", length, i+4)
		}
		value := input[i+4 : i+4+length]

		tlv := TLV{Tag: tag, Value: value}

		if containerTags[tag] {
			if children, err := Decode(value); err == nil {
				var consumed int
				for _, c := range children {
					consumed += len(c.Encoded())
				}
				if consumed == len(value) && len(children) > 0 {
					tlv.Children = children
				}
			}
		}

		tlvs = append(tlvs, tlv)
		i += 4 + length
	}
	return tlvs, nil
}

func DecodeRaw(input string) ([]TLV, error) {
	var tlvs []TLV
	for i := 0; i < len(input); {
		if i+4 > len(input) {
			return nil, fmt.Errorf("unexpected end at position %d", i)
		}
		tag := input[i : i+2]
		lengthStr := input[i+2 : i+4]

		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, fmt.Errorf("invalid length '%s' at position %d", lengthStr, i+2)
		}

		if i+4+length > len(input) {
			return nil, fmt.Errorf("value length %d exceeds remaining input at position %d", length, i+4)
		}
		value := input[i+4 : i+4+length]

		tlvs = append(tlvs, TLV{Tag: tag, Value: value})
		i += 4 + length
	}
	return tlvs, nil
}

func FormatTLVs(tlvs []TLV, indent string) string {
	var sb strings.Builder
	for _, tlv := range tlvs {
		name := tagName(tlv.Tag)
		sb.WriteString(fmt.Sprintf("%s%s [len=%d]", indent, name, len(tlv.Value)))
		if len(tlv.Children) > 0 {
			sb.WriteString("\n")
			sb.WriteString(FormatTLVs(tlv.Children, indent+"    "))
		} else if tlv.Tag == "63" {
			sb.WriteString(fmt.Sprintf(" = %s\n", tlv.Value))
		} else {
			sb.WriteString(fmt.Sprintf(" = \"%s\"\n", tlv.Value))
		}
	}
	return sb.String()
}

func FormatTLVsRaw(tlvs []TLV, indent string) string {
	var sb strings.Builder
	for _, tlv := range tlvs {
		sb.WriteString(fmt.Sprintf("%sTag %s  Len %d  Val=", indent, tlv.Tag, len(tlv.Value)))
		if len(tlv.Value) > 40 {
			sb.WriteString(fmt.Sprintf("%s...(%d bytes)", tlv.Value[:40], len(tlv.Value)))
		} else {
			sb.WriteString(tlv.Value)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func FindTLV(tlvs []TLV, tag string) *TLV {
	for i := range tlvs {
		if tlvs[i].Tag == tag {
			return &tlvs[i]
		}
	}
	return nil
}
