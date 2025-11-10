package payment

import (
	"embed"
	"html/template"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

//go:embed payment.html
var htmlFS embed.FS

type PaymentPageData struct {
	OrderID      string
	OrderIDShort string
	ReturnURL    string
	Price        string
}

func generateMockPrice() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	prices := []string{"999.99", "1499.99", "1999.99", "249.99", "299.99", "499.99"}
	return prices[rand.Intn(len(prices))]
}

func formatPrice(s string) string { // TODO: пофиксить чтобы работала запятая нормально
	var clean strings.Builder
	for _, c := range s {
		if c >= '0' && c <= '9' || c == '.' {
			clean.WriteRune(c)
		}
	}
	num := clean.String()

	parts := strings.Split(num, ".")
	intPart := parts[0]
	decPart := "00"
	if len(parts) > 1 {
		decPart = parts[1]
		if len(decPart) > 2 {
			decPart = decPart[:2]
		} else if len(decPart) == 1 {
			decPart += "0"
		}
	}

	var formattedInt strings.Builder
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			formattedInt.WriteRune(' ')
		}
		formattedInt.WriteRune(c)
	}

	return formattedInt.String() + "," + decPart
}

func RenderPaymentPage(w http.ResponseWriter, data *PaymentPageData) error {
	tmpl, err := template.ParseFS(htmlFS, "payment.html")
	if err != nil {
		return err
	}

	if data.Price == "" {
		data.Price = generateMockPrice()
	} else {
		data.Price = formatPrice(data.Price)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}
