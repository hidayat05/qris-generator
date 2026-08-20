package main

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
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
	a.Settings().SetTheme(&neumorphicTheme{})
	w := a.NewWindow("QRIS Generator")
	w.Resize(fyne.NewSize(2560, 1600))

	tabs := container.NewAppTabs(
		container.NewTabItem("Generate QRIS", generateTab(w)),
		container.NewTabItem("Parse QRIS", parseTab(w)),
		container.NewTabItem("TLV Tree", tlvTreeTab(w)),
	)

	w.SetContent(tabs)
	w.ShowAndRun()
}

func generateTab(w fyne.Window) fyne.CanvasObject {
	importEntry := widget.NewEntry()
	importEntry.PlaceHolder = "Paste raw QRIS string to load details..."

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

	fixedFee := widget.NewEntry()
	fixedFee.PlaceHolder = "e.g. 5000"
	fixedFee.Disable()

	percentFee := widget.NewEntry()
	percentFee.PlaceHolder = "e.g. 2.50"
	percentFee.Disable()

	tipTypes := []string{"None", "Prompt for Tip (01)", "Fixed Fee (02)", "Percentage Fee (03)"}
	tipSelector := widget.NewSelect(tipTypes, func(selected string) {
		switch selected {
		case "Prompt for Tip (01)":
			fixedFee.Disable()
			fixedFee.SetText("")
			percentFee.Disable()
			percentFee.SetText("")
		case "Fixed Fee (02)":
			fixedFee.Enable()
			percentFee.Disable()
			percentFee.SetText("")
		case "Percentage Fee (03)":
			fixedFee.Disable()
			fixedFee.SetText("")
			percentFee.Enable()
		default: // "None"
			fixedFee.Disable()
			fixedFee.SetText("")
			percentFee.Disable()
			percentFee.SetText("")
		}
	})
	tipSelector.SetSelected("None")

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

	qrSectionBackground := canvas.NewRectangle(color.White)
	qrSectionBackground.Hide()

	qrSection := container.NewStack(qrSectionBackground, container.NewCenter(qrCanvas))

	statusLabel := widget.NewLabel("")

	importBtn := widget.NewButton("Load / Import", func() {
		raw := strings.TrimSpace(importEntry.Text)
		if raw == "" {
			dialog.ShowInformation("Import Error", "Please paste a raw QRIS string first", w)
			return
		}
		data, err := qris.ParseQRIS(raw)
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to parse raw QRIS: %w", err), w)
			return
		}

		merchantName.SetText(data.MerchantName)
		merchantCity.SetText(data.MerchantCity)
		merchantID.SetText(data.MerchantAccountInfo.MerchantID)
		if data.MerchantCategoryCode != "" {
			mcc.SetText(data.MerchantCategoryCode)
		} else {
			mcc.SetText("5211")
		}

		if data.PointOfInitiationMethod == "12" {
			typeSelector.SetSelected("Dynamic (with amount)")
			amount.SetText(data.TransactionAmount)
			amount.Enable()
		} else {
			typeSelector.SetSelected("Static (no amount)")
			amount.SetText("")
			amount.Disable()
		}

		if data.TransactionCurrency != "" {
			currency.SetSelected(currencyName(data.TransactionCurrency))
		}
		if data.CountryCode != "" {
			countryCode.SetSelected(countryName(data.CountryCode))
		}
		postalCode.SetText(data.PostalCode)

		if data.TipIndicator != "" {
			switch data.TipIndicator {
			case "01":
				tipSelector.SetSelected("Prompt for Tip (01)")
			case "02":
				tipSelector.SetSelected("Fixed Fee (02)")
				fixedFee.SetText(data.ConvenienceFeeFixed)
			case "03":
				tipSelector.SetSelected("Percentage Fee (03)")
				percentFee.SetText(data.ConvenienceFeePercentage)
			default:
				tipSelector.SetSelected("None")
			}
		} else {
			tipSelector.SetSelected("None")
		}

		statusLabel.SetText("Imported QRIS details successfully!")
	})

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

		switch tipSelector.Selected {
		case "Prompt for Tip (01)":
			data.TipIndicator = "01"
		case "Fixed Fee (02)":
			data.TipIndicator = "02"
			data.ConvenienceFeeFixed = strings.TrimSpace(fixedFee.Text)
			if data.ConvenienceFeeFixed == "" {
				dialog.ShowInformation("Validation", "Fixed Fee is required", w)
				return
			}
		case "Percentage Fee (03)":
			data.TipIndicator = "03"
			data.ConvenienceFeePercentage = strings.TrimSpace(percentFee.Text)
			if data.ConvenienceFeePercentage == "" {
				dialog.ShowInformation("Validation", "Percentage Fee is required", w)
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
		qr.BackgroundColor = color.White
		qr.ForegroundColor = color.Black
		img := qr.Image(280)
		qrCanvas.Image = img
		qrCanvas.Refresh()
		qrSectionBackground.Show()
		qrSectionBackground.Refresh()
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
		tipSelector.SetSelected("None")
		fixedFee.SetText("")
		percentFee.SetText("")
		importEntry.SetText("")
		resultEntry.SetText("")
		qrCanvas.Image = nil
		qrCanvas.Refresh()
		qrSectionBackground.Hide()
		qrSectionBackground.Refresh()
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
		&widget.FormItem{Text: "Tip / Convenience Fee", Widget: tipSelector},
		&widget.FormItem{Text: "Convenience Fee (Fixed)", Widget: fixedFee},
		&widget.FormItem{Text: "Convenience Fee (%)", Widget: percentFee},
	)

	importContainer := container.NewBorder(
		nil, nil, nil, importBtn,
		importEntry,
	)

	leftSide := container.NewVBox(
		widget.NewLabelWithStyle("Import Existing QRIS String", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		importContainer,
		widget.NewSeparator(),
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
			qrSection,
		),
	)

	leftScroll := container.NewVScroll(leftSide)
	split := container.NewHSplit(leftScroll, rightSide)
	split.Offset = 0.48

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
		if data.TipIndicator != "" {
			var tipText string
			switch data.TipIndicator {
			case "01":
				tipText = "Prompt for Tip (01)"
			case "02":
				tipText = "Fixed Convenience Fee (02)"
			case "03":
				tipText = "Percentage Convenience Fee (03)"
			default:
				tipText = "Unknown Indicator (" + data.TipIndicator + ")"
			}
			sb.WriteString(fmt.Sprintf("Tip Indicator: %s\n", tipText))
		}
		if data.ConvenienceFeeFixed != "" {
			sb.WriteString(fmt.Sprintf("Convenience Fee (Fixed): %s\n", data.ConvenienceFeeFixed))
		}
		if data.ConvenienceFeePercentage != "" {
			sb.WriteString(fmt.Sprintf("Convenience Fee (%%): %s%%\n", data.ConvenienceFeePercentage))
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

type neumorphicTheme struct{}

var _ fyne.Theme = (*neumorphicTheme)(nil)

func (n *neumorphicTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if variant == theme.VariantDark {
		switch name {
		case theme.ColorNameBackground:
			return color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 0xFF} // soft pure dark gray
		case theme.ColorNameHeaderBackground:
			return color.RGBA{R: 0x15, G: 0x15, B: 0x15, A: 0xFF}
		case theme.ColorNameButton:
			return color.RGBA{R: 0x2A, G: 0x2A, B: 0x2A, A: 0xFF} // slightly elevated gray
		case theme.ColorNameDisabledButton:
			return color.RGBA{R: 0x1E, G: 0x1E, B: 0x1E, A: 0xFF}
		case theme.ColorNameInputBackground:
			return color.RGBA{R: 0x14, G: 0x14, B: 0x14, A: 0xFF} // inset inputs
		case theme.ColorNameInputBorder:
			return color.RGBA{R: 0x24, G: 0x24, B: 0x24, A: 0xFF}
		case theme.ColorNameForeground:
			return color.RGBA{R: 0xEA, G: 0xEA, B: 0xEA, A: 0xFF} // off-white text
		case theme.ColorNamePlaceHolder:
			return color.RGBA{R: 0x76, G: 0x76, B: 0x76, A: 0xFF}
		case theme.ColorNamePrimary:
			return color.RGBA{R: 0x81, G: 0x8C, B: 0xF8, A: 0xFF} // Muted indigo
		case theme.ColorNameHover:
			return color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF}
		case theme.ColorNamePressed:
			return color.RGBA{R: 0x18, G: 0x18, B: 0x18, A: 0xFF}
		case theme.ColorNameFocus:
			return color.RGBA{R: 0x63, G: 0x66, B: 0xF1, A: 0xFF}
		case theme.ColorNameSelection:
			return color.RGBA{R: 0x31, G: 0x2E, B: 0x81, A: 0x80}
		case theme.ColorNameSeparator:
			return color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x14} // ultra-lightweight white separator (8% opacity)
		case theme.ColorNameShadow:
			return color.RGBA{R: 0x0A, G: 0x0A, B: 0x0A, A: 0xE0}
		}
	} else {
		// VariantLight
		switch name {
		case theme.ColorNameBackground:
			return color.RGBA{R: 0xE0, G: 0xE5, B: 0xEC, A: 0xFF} // Neumorphic background
		case theme.ColorNameHeaderBackground:
			return color.RGBA{R: 0xD2, G: 0xD9, B: 0xE4, A: 0xFF}
		case theme.ColorNameButton:
			return color.RGBA{R: 0xE0, G: 0xE5, B: 0xEC, A: 0xFF} // matching background for neumorphism
		case theme.ColorNameDisabledButton:
			return color.RGBA{R: 0xD2, G: 0xD9, B: 0xE4, A: 0xFF}
		case theme.ColorNameInputBackground:
			return color.RGBA{R: 0xD2, G: 0xD9, B: 0xE4, A: 0xFF} // inset inputs
		case theme.ColorNameInputBorder:
			return color.RGBA{R: 0xC3, G: 0xCD, B: 0xD8, A: 0xFF}
		case theme.ColorNameForeground:
			return color.RGBA{R: 0x2D, G: 0x37, B: 0x48, A: 0xFF} // Slate-800
		case theme.ColorNamePlaceHolder:
			return color.RGBA{R: 0x71, G: 0x80, B: 0x96, A: 0xFF}
		case theme.ColorNamePrimary:
			return color.RGBA{R: 0x4F, G: 0x46, B: 0xE5, A: 0xFF} // Indigo-600
		case theme.ColorNameHover:
			return color.RGBA{R: 0xEC, G: 0xF0, B: 0xF5, A: 0xFF} // hover highlight
		case theme.ColorNamePressed:
			return color.RGBA{R: 0xC5, G: 0xCE, B: 0xDB, A: 0xFF} // pressed shadow
		case theme.ColorNameFocus:
			return color.RGBA{R: 0x63, G: 0x66, B: 0xF1, A: 0xFF}
		case theme.ColorNameSelection:
			return color.RGBA{R: 0xC7, G: 0xD2, B: 0xFE, A: 0xFF}
		case theme.ColorNameSeparator:
			return color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0x1A} // ultra-lightweight black separator (10% opacity)
		case theme.ColorNameShadow:
			return color.RGBA{R: 0x9B, G: 0xAD, B: 0xC6, A: 0x80} // soft neumorphic shadow
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (n *neumorphicTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInputRadius:
		return 12 // elegant rounded inputs
	case theme.SizeNameSelectionRadius:
		return 8
	case theme.SizeNamePadding:
		return 12 // cleaner padding
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameText:
		return 14 // highly readable body text
	case theme.SizeNameHeadingText:
		return 18
	case theme.SizeNameSeparatorThickness:
		return 1 // ultra-thin separator thickness
	}
	return theme.DefaultTheme().Size(name)
}

func (n *neumorphicTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (n *neumorphicTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
