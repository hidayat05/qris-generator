package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	qrcode "github.com/skip2/go-qrcode"

	"qris-generator/qris"
)

func currencyCode(selected string) string {
	if idx := strings.Index(selected, " - "); idx != -1 {
		return selected[:idx]
	}
	return selected
}

func currencyName(code string) string {
	currencies := []string{
		"036 - AUD (Australian Dollar)",
		"124 - CAD (Canadian Dollar)",
		"156 - CNY (Chinese Yuan)",
		"344 - HKD (Hong Kong Dollar)",
		"360 - IDR (Indonesian Rupiah)",
		"392 - JPY (Japanese Yen)",
		"410 - KRW (South Korean Won)",
		"458 - MYR (Malaysian Ringgit)",
		"554 - NZD (New Zealand Dollar)",
		"608 - PHP (Philippine Peso)",
		"682 - SAR (Saudi Riyal)",
		"702 - SGD (Singapore Dollar)",
		"704 - VND (Vietnamese Dong)",
		"764 - THB (Thai Baht)",
		"784 - AED (UAE Dirham)",
		"826 - GBP (British Pound)",
		"840 - USD (US Dollar)",
		"978 - EUR (Euro)",
	}
	for _, c := range currencies {
		if strings.HasPrefix(c, code+" - ") {
			return c
		}
	}
	return code
}

func countryName(code string) string {
	countries := []string{
		"ID - Indonesia",
		"AE - United Arab Emirates",
		"AU - Australia",
		"BN - Brunei Darussalam",
		"CN - China",
		"GB - United Kingdom",
		"HK - Hong Kong",
		"JP - Japan",
		"KR - South Korea",
		"MY - Malaysia",
		"NZ - New Zealand",
		"PH - Philippines",
		"QA - Qatar",
		"SA - Saudi Arabia",
		"SG - Singapore",
		"TH - Thailand",
		"US - United States",
		"VN - Vietnam",
	}
	for _, c := range countries {
		if strings.HasPrefix(c, code+" - ") {
			return c
		}
	}
	return code
}

func main() {
	a := app.New()
	w := a.NewWindow("QRIS Generator")
	w.Resize(fyne.NewSize(900, 700))

	tabs := container.NewAppTabs(
		container.NewTabItem("Generate QRIS", generateTab(w)),
		container.NewTabItem("Parse QRIS", parseTab(w)),
		container.NewTabItem("TLV Tree", tlvTreeTab(w)),
	)

	w.SetContent(tabs)
	w.ShowAndRun()
}

func generateTab(w fyne.Window) fyne.CanvasObject {
	merchantName := widget.NewEntry()
	merchantName.PlaceHolder = "e.g. TOKO MAKMUR"

	merchantCity := widget.NewEntry()
	merchantCity.PlaceHolder = "e.g. JAKARTA"

	merchantID := widget.NewEntry()
	merchantID.PlaceHolder = "e.g. 9876543210"

	mcc := widget.NewEntry()
	mcc.SetText("5211")

	amount := widget.NewEntry()
	amount.PlaceHolder = "e.g. 50000"
	amount.Disable()

	countries := []string{
		"ID - Indonesia",
		"AE - United Arab Emirates",
		"AU - Australia",
		"BN - Brunei Darussalam",
		"CN - China",
		"GB - United Kingdom",
		"HK - Hong Kong",
		"JP - Japan",
		"KR - South Korea",
		"MY - Malaysia",
		"NZ - New Zealand",
		"PH - Philippines",
		"QA - Qatar",
		"SA - Saudi Arabia",
		"SG - Singapore",
		"TH - Thailand",
		"US - United States",
		"VN - Vietnam",
	}
	countryCode := widget.NewSelect(countries, nil)
	countryCode.SetSelected("ID - Indonesia")

	currencies := []string{
		"360 - IDR (Indonesian Rupiah)",
		"036 - AUD (Australian Dollar)",
		"124 - CAD (Canadian Dollar)",
		"156 - CNY (Chinese Yuan)",
		"344 - HKD (Hong Kong Dollar)",
		"392 - JPY (Japanese Yen)",
		"410 - KRW (South Korean Won)",
		"458 - MYR (Malaysian Ringgit)",
		"554 - NZD (New Zealand Dollar)",
		"608 - PHP (Philippine Peso)",
		"682 - SAR (Saudi Riyal)",
		"702 - SGD (Singapore Dollar)",
		"704 - VND (Vietnamese Dong)",
		"764 - THB (Thai Baht)",
		"784 - AED (UAE Dirham)",
		"826 - GBP (British Pound)",
		"840 - USD (US Dollar)",
		"978 - EUR (Euro)",
	}
	currency := widget.NewSelect(currencies, nil)
	currency.SetSelected("360 - IDR (Indonesian Rupiah)")

	postalCode := widget.NewEntry()
	postalCode.PlaceHolder = "Optional"

	typeSelector := widget.NewSelect([]string{"Static (no amount)", "Dynamic (with amount)"}, func(s string) {
		if s == "Dynamic (with amount)" {
			amount.Enable()
		} else {
			amount.Disable()
			amount.SetText("")
		}
	})
	typeSelector.SetSelected("Static (no amount)")

	resultEntry := widget.NewMultiLineEntry()
	resultEntry.Wrapping = fyne.TextWrapBreak
	resultEntry.Disable()

	qrCanvas := canvas.NewImageFromImage(nil)
	qrCanvas.SetMinSize(fyne.NewSize(280, 280))
	qrCanvas.FillMode = canvas.ImageFillContain

	statusLabel := widget.NewLabel("")

	generateBtn := widget.NewButton("Generate QRIS", func() {
		if strings.TrimSpace(merchantName.Text) == "" {
			dialog.ShowInformation("Validation", "Merchant Name is required", w)
			return
		}
		if strings.TrimSpace(merchantID.Text) == "" {
			dialog.ShowInformation("Validation", "Merchant ID is required", w)
			return
		}

		data := &qris.QRISData{
			PayloadFormatIndicator: "01",
			MerchantCategoryCode:   mcc.Text,
			TransactionCurrency:    currencyCode(currency.Selected),
			CountryCode:            currencyCode(countryCode.Selected),
			MerchantName:           merchantName.Text,
			MerchantCity:           merchantCity.Text,
			PostalCode:             postalCode.Text,
			MerchantAccountInfo: qris.MerchantAccountInfo{
				GUID:       "54",
				MerchantID: merchantID.Text,
			},
		}

		if typeSelector.Selected == "Static (no amount)" {
			data.PointOfInitiationMethod = "11"
		} else {
			data.PointOfInitiationMethod = "12"
			data.TransactionAmount = strings.TrimSpace(amount.Text)
			if data.TransactionAmount == "" {
				dialog.ShowInformation("Validation", "Amount is required for dynamic QR", w)
				return
			}
		}

		raw := qris.GenerateQRIS(data)
		resultEntry.SetText(raw)
		statusLabel.SetText("")

		qr, err := qrcode.New(raw, qrcode.Medium)
		if err != nil {
			dialog.ShowError(fmt.Errorf("QR code generation failed: %w", err), w)
			return
		}
		qr.DisableBorder = true
		img := qr.Image(280)
		qrCanvas.Image = img
		qrCanvas.Refresh()
		statusLabel.SetText("QRIS generated successfully!")
	})

	copyBtn := widget.NewButton("Copy to Clipboard", func() {
		if resultEntry.Text == "" {
			return
		}
		w.Clipboard().SetContent(resultEntry.Text)
		statusLabel.SetText("Copied to clipboard!")
	})

	clearBtn := widget.NewButton("Clear", func() {
		merchantName.SetText("")
		merchantCity.SetText("")
		merchantID.SetText("")
		mcc.SetText("5211")
		amount.SetText("")
		countryCode.SetSelected("ID - Indonesia")
		currency.SetSelected("360 - IDR (Indonesian Rupiah)")
		postalCode.SetText("")
		resultEntry.SetText("")
		qrCanvas.Image = nil
		qrCanvas.Refresh()
		statusLabel.SetText("")
	})

	form := widget.NewForm(
		widget.NewFormItem("Merchant Name *", merchantName),
		widget.NewFormItem("Merchant City", merchantCity),
		widget.NewFormItem("Merchant ID *", merchantID),
		widget.NewFormItem("MCC", mcc),
		widget.NewFormItem("Currency", currency),
		widget.NewFormItem("Country", countryCode),
		widget.NewFormItem("Postal Code", postalCode),
		&widget.FormItem{Text: "QR Type", Widget: typeSelector},
		&widget.FormItem{Text: "Amount", Widget: amount},
	)

	leftSide := container.NewVBox(
		widget.NewLabelWithStyle("QRIS Details", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		form,
		container.NewHBox(generateBtn, copyBtn, clearBtn),
		statusLabel,
	)

	rightTop := widget.NewLabelWithStyle("Generated QRIS Raw String", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	rightSide := container.NewBorder(
		rightTop, nil, nil, nil,
		container.NewVSplit(
			resultEntry,
			container.NewCenter(qrCanvas),
		),
	)

	split := container.NewHSplit(leftSide, rightSide)
	split.Offset = 0.45

	return split
}

func parseTab(w fyne.Window) fyne.CanvasObject {
	input := widget.NewMultiLineEntry()
	input.PlaceHolder = "Paste raw QRIS string here..."
	input.Wrapping = fyne.TextWrapBreak

	resultLabel := widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord

	statusLabel := widget.NewLabel("")

	parseBtn := widget.NewButton("Parse", func() {
		raw := strings.TrimSpace(input.Text)
		if raw == "" {
			dialog.ShowInformation("Error", "Please enter a QRIS string first", w)
			return
		}

		data, err := qris.ParseQRIS(raw)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Parse failed: %w", err), w)
			return
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Payload Format Indicator: %s\n", data.PayloadFormatIndicator))
		sb.WriteString(fmt.Sprintf("Point of Initiation: %s\n", data.PointOfInitiationMethod))
		sb.WriteString(fmt.Sprintf("Merchant Name: %s\n", data.MerchantName))
		sb.WriteString(fmt.Sprintf("Merchant City: %s\n", data.MerchantCity))
		sb.WriteString(fmt.Sprintf("MCC: %s\n", data.MerchantCategoryCode))
		sb.WriteString(fmt.Sprintf("Currency: %s\n", currencyName(data.TransactionCurrency)))
		sb.WriteString(fmt.Sprintf("Country: %s\n", countryName(data.CountryCode)))

		if data.PostalCode != "" {
			sb.WriteString(fmt.Sprintf("Postal Code: %s\n", data.PostalCode))
		}
		if data.TransactionAmount != "" {
			sb.WriteString(fmt.Sprintf("Amount: %s\n", data.TransactionAmount))
		}
		sb.WriteString(fmt.Sprintf("GUID: %s\n", data.MerchantAccountInfo.GUID))
		sb.WriteString(fmt.Sprintf("PAN: %s\n", data.MerchantAccountInfo.PAN))
		sb.WriteString(fmt.Sprintf("Merchant ID: %s\n", data.MerchantAccountInfo.MerchantID))

		if data.AdditionalData.BillNumber != "" {
			sb.WriteString(fmt.Sprintf("Bill Number: %s\n", data.AdditionalData.BillNumber))
		}
		if data.AdditionalData.MobileNumber != "" {
			sb.WriteString(fmt.Sprintf("Mobile: %s\n", data.AdditionalData.MobileNumber))
		}
		if data.AdditionalData.StoreLabel != "" {
			sb.WriteString(fmt.Sprintf("Store Label: %s\n", data.AdditionalData.StoreLabel))
		}
		if data.AdditionalData.LoyaltyNumber != "" {
			sb.WriteString(fmt.Sprintf("Loyalty: %s\n", data.AdditionalData.LoyaltyNumber))
		}
		if data.AdditionalData.TaxLabel != "" {
			sb.WriteString(fmt.Sprintf("Tax: %s\n", data.AdditionalData.TaxLabel))
		}

		valid := qris.VerifyCRC(raw)
		sb.WriteString(fmt.Sprintf("\nCRC Valid: %v", valid))
		if !valid {
			sb.WriteString(" \u26A0 WARNING: QRIS string may be corrupted or invalid")
		}

		resultLabel.SetText(sb.String())
		statusLabel.SetText("Parse complete")
	})

	clearBtn := widget.NewButton("Clear", func() {
		input.SetText("")
		resultLabel.SetText("")
		statusLabel.SetText("")
	})

	topContainer := container.NewVBox(
		widget.NewLabelWithStyle("Enter Raw QRIS String", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		input,
		container.NewHBox(parseBtn, clearBtn),
		statusLabel,
	)

	resultScroll := container.NewVScroll(resultLabel)
	resultScroll.SetMinSize(fyne.NewSize(0, 300))

	return container.NewVSplit(topContainer, resultScroll)
}

func tlvTreeTab(w fyne.Window) fyne.CanvasObject {
	input := widget.NewMultiLineEntry()
	input.PlaceHolder = "Paste raw QRIS string here..."
	input.Wrapping = fyne.TextWrapBreak

	resultLabel := widget.NewLabel("")
	resultLabel.Wrapping = fyne.TextWrapWord
	resultLabel.TextStyle = fyne.TextStyle{Monospace: true}

	statusLabel := widget.NewLabel("")

	parseBtn := widget.NewButton("Formatted Tree", func() {
		raw := strings.TrimSpace(input.Text)
		if raw == "" {
			dialog.ShowInformation("Error", "Please enter a QRIS string first", w)
			return
		}

		tlvs, err := qris.Decode(raw)
		if err != nil {
			dialog.ShowError(fmt.Errorf("decode failed: %w", err), w)
			return
		}

		resultLabel.SetText(qris.FormatTLVs(tlvs, ""))
		statusLabel.SetText("Formatted tree displayed")
	})

	rawBtn := widget.NewButton("Raw TLV", func() {
		raw := strings.TrimSpace(input.Text)
		if raw == "" {
			dialog.ShowInformation("Error", "Please enter a QRIS string first", w)
			return
		}

		tlvs, err := qris.DecodeRaw(raw)
		if err != nil {
			dialog.ShowError(fmt.Errorf("raw decode failed: %w", err), w)
			return
		}

		var sb strings.Builder
		for _, tlv := range tlvs {
			sb.WriteString(fmt.Sprintf("Tag %s  Len %d  Val=%q\n", tlv.Tag, len(tlv.Value), tlv.Value))
		}
		resultLabel.SetText(sb.String())
		statusLabel.SetText("Raw TLV displayed")
	})

	clearBtn := widget.NewButton("Clear", func() {
		input.SetText("")
		resultLabel.SetText("")
		statusLabel.SetText("")
	})

	topContainer := container.NewVBox(
		widget.NewLabelWithStyle("TLV Tree Viewer", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Paste a QRIS string to inspect its TLV structure"),
		input,
		container.NewHBox(parseBtn, rawBtn, clearBtn),
		statusLabel,
	)

	resultScroll := container.NewVScroll(resultLabel)
	resultScroll.SetMinSize(fyne.NewSize(0, 300))

	return container.NewVSplit(topContainer, resultScroll)
}
