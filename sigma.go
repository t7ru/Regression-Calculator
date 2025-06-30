package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func formatFloatSmart(val float64, useRounding bool) string {
	if useRounding {
		roundedVal := math.Round(val*100) / 100
		return strconv.FormatFloat(roundedVal, 'f', -1, 64)
	} else {
		return strconv.FormatFloat(val, 'f', -1, 64)
	}
}

func calculateAndGenerateOutput(x, y []float64, regressionType string, useRounding bool) (string, error) {
	var outputText strings.Builder

	if len(x) == 0 || len(y) == 0 {
		return "", fmt.Errorf("x or y values are empty")
	}
	if len(x) != len(y) {
		return "", fmt.Errorf("the number of x and y values must be the same")
	}

	n := len(x)
	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for i := 0; i < n; i++ {
		xi := x[i]
		yi := y[i]
		sumX += xi
		sumY += yi
		sumX2 += xi * xi
		sumY2 += yi * yi
		sumXY += xi * yi
	}

	outputText.WriteString(fmt.Sprintf("%-7s %-10s %-10s %-12s %-12s %-12s\n", "Index", "X", "Y", "X^2", "Y^2", "XY"))
	for i := 0; i < n; i++ {
		xi := x[i]
		yi := y[i]
		outputText.WriteString(fmt.Sprintf("%-7d %-10s %-10s %-12s %-12s %-12s\n", i+1,
			formatFloatSmart(xi, useRounding),
			formatFloatSmart(yi, useRounding),
			formatFloatSmart(xi*xi, useRounding),
			formatFloatSmart(yi*yi, useRounding),
			formatFloatSmart(xi*yi, useRounding)))
	}

	outputText.WriteString(fmt.Sprintf("\nΣx = %s\n", formatFloatSmart(sumX, useRounding)))
	outputText.WriteString(fmt.Sprintf("Σy = %s\n", formatFloatSmart(sumY, useRounding)))
	outputText.WriteString(fmt.Sprintf("Σx² = %s\n", formatFloatSmart(sumX2, useRounding)))
	outputText.WriteString(fmt.Sprintf("Σy² = %s\n", formatFloatSmart(sumY2, useRounding)))
	outputText.WriteString(fmt.Sprintf("Σxy = %s\n", formatFloatSmart(sumXY, useRounding)))
	outputText.WriteString(fmt.Sprintf("n = %d\n", n))

	numerator := float64(n)*sumXY - sumX*sumY
	termX := float64(n)*sumX2 - sumX*sumX
	termY := float64(n)*sumY2 - sumY*sumY
	rDenominator := termX * termY

	if rDenominator != 0 {
		rVal := numerator / math.Sqrt(rDenominator)
		outputText.WriteString(fmt.Sprintf("r = %.4f\n", rVal))
		outputText.WriteString(fmt.Sprintf("r² = %.4f\n", rVal*rVal))

		if rVal > 0 {
			outputText.WriteString("The correlation is positive.\n")
		} else if rVal < 0 {
			outputText.WriteString("The correlation is negative.\n")
		} else {
			outputText.WriteString("There is no pos/neg correlation.\n")
		}

		absR := math.Abs(rVal)
		switch {
		case absR == 1.0:
			outputText.WriteString("Perfect correlation.\n")
		case absR >= 0.7:
			outputText.WriteString("Strong correlation.\n")
		case absR >= 0.5:
			outputText.WriteString("Moderate correlation.\n")
		case absR >= 0.3:
			outputText.WriteString("Weak correlation.\n")
		default:
			outputText.WriteString("There is no spectrum correlation.\n")
		}
	} else {
		outputText.WriteString("r is undefined (denominator for r is zero).\n")
	}

	var meanY, stdDevY float64
	for _, yi := range y {
		meanY += yi
	}
	meanY /= float64(n)

	for _, yi := range y {
		stdDevY += (yi - meanY) * (yi - meanY)
	}
	stdDevY = math.Sqrt(stdDevY / float64(n))

	outliers := []int{}
	threshold := 2.0
	for i, yi := range y {
		if math.Abs(yi-meanY) > threshold*stdDevY {
			outliers = append(outliers, i)
		}
	}

	outputText.WriteString(fmt.Sprintf("\nMean of Y: %s\n", formatFloatSmart(meanY, useRounding)))
	outputText.WriteString(fmt.Sprintf("Standard Deviation of Y: %s\n", formatFloatSmart(stdDevY, useRounding)))
	if len(outliers) > 0 {
		outputText.WriteString("Outliers detected at the following indices (0-based):\n")
		for _, idx := range outliers {
			outputText.WriteString(fmt.Sprintf("Index %d: X = %s, Y = %s\n", idx, formatFloatSmart(x[idx], useRounding), formatFloatSmart(y[idx], useRounding)))
		}
	} else {
		outputText.WriteString("No outliers detected.\n")
	}

	if regressionType == "quadratic" {
		sumX3 := 0.0
		sumX4 := 0.0
		sumX2Y := 0.0

		for i := 0; i < n; i++ {
			xi := x[i]
			yi := y[i]
			sumX3 += xi * xi * xi
			sumX4 += xi * xi * xi * xi
			sumX2Y += xi * xi * yi
		}

		det := (sumX4*(sumX2*float64(n)-sumX*sumX) -
			sumX3*(sumX3*float64(n)-sumX*sumX2) +
			sumX2*(sumX3*sumX-sumX2*sumX2))

		if det != 0 {
			a := (sumX2Y*(sumX2*float64(n)-sumX*sumX) -
				sumXY*(sumX3*float64(n)-sumX*sumX2) +
				sumY*(sumX3*sumX-sumX2*sumX2)) / det

			b := (sumX4*(sumXY*float64(n)-sumY*sumX) -
				sumX3*(sumX2Y*float64(n)-sumY*sumX2) +
				sumX2*(sumX2Y*sumX-sumXY*sumX2)) / det

			c := (sumX4*(sumX2*sumY-sumX*sumXY) -
				sumX3*(sumX3*sumY-sumX*sumX2Y) +
				sumX2*(sumX3*sumXY-sumX2*sumX2Y)) / det

			vertexX := -b / (2 * a)
			vertexY := a*vertexX*vertexX + b*vertexX + c

			vertexType := "minimum"
			if a < 0 {
				vertexType = "maximum"
			}

			outputText.WriteString(fmt.Sprintf("\nQuadratic Trendline equation: y = %sx² + %sx + %s\n",
				formatFloatSmart(a, useRounding), formatFloatSmart(b, useRounding), formatFloatSmart(c, useRounding)))

			outputText.WriteString(fmt.Sprintf("Vertex: (%s, %s) - This is a %s\n",
				formatFloatSmart(vertexX, useRounding), formatFloatSmart(vertexY, useRounding), vertexType))
		} else {
			outputText.WriteString("\nCould not calculate quadratic regression (determinant is zero).\n")
		}
	} else if regressionType == "exponential" {
		for i, yi := range y {
			if yi <= 0 {
				outputText.WriteString(fmt.Sprintf("\nCannot perform exponential regression: y value at index %d is not positive (%s).\n", i, formatFloatSmart(yi, useRounding)))
				return outputText.String(), nil
			}
		}

		shiftedX := make([]float64, n)
		minX := x[0]
		for _, xi := range x {
			if xi < minX {
				minX = xi
			}
		}
		for i := range x {
			shiftedX[i] = x[i] - minX
		}

		lnY := make([]float64, n)
		for i, yi := range y {
			lnY[i] = math.Log(yi)
		}

		sumLnY := 0.0
		sumXLnY := 0.0
		sumShiftedX := 0.0
		sumShiftedX2 := 0.0
		for i := 0; i < n; i++ {
			sumLnY += lnY[i]
			sumXLnY += shiftedX[i] * lnY[i]
			sumShiftedX += shiftedX[i]
			sumShiftedX2 += shiftedX[i] * shiftedX[i]
		}

		termX := float64(n)*sumShiftedX2 - sumShiftedX*sumShiftedX
		if termX == 0 {
			outputText.WriteString("\nCannot perform exponential regression: denominator for b is zero (all X values may be the same).\n")
			return outputText.String(), nil
		}
		b := (float64(n)*sumXLnY - sumShiftedX*sumLnY) / termX
		lnA := (sumLnY - b*sumShiftedX) / float64(n)
		a := math.Exp(lnA)

		if a == 0 || math.IsNaN(a) || math.IsInf(a, 0) {
			outputText.WriteString("\nCould not calculate exponential regression: invalid coefficient a (possibly due to data scale or input values).\n")
			return outputText.String(), nil
		}

		outputText.WriteString(fmt.Sprintf("\nExponential equation: y = %s*e^(%s*(x-%s))\n",
			formatFloatSmart(a, useRounding), formatFloatSmart(b, useRounding), formatFloatSmart(minX, useRounding)))

		outputText.WriteString(fmt.Sprintf("Solving for x: x = %s + ln(y/%s)/%s\n",
			formatFloatSmart(minX, useRounding), formatFloatSmart(a, useRounding), formatFloatSmart(b, useRounding)))

		sumLnY2 := 0.0
		for i := 0; i < n; i++ {
			sumLnY2 += lnY[i] * lnY[i]
		}

		numeratorLn := float64(n)*sumXLnY - sumShiftedX*sumLnY
		termLnY := float64(n)*sumLnY2 - sumLnY*sumLnY
		rDenominatorLn := termX * termLnY

		if rDenominatorLn > 0 {
			rLn := numeratorLn / math.Sqrt(rDenominatorLn)
			outputText.WriteString(fmt.Sprintf("r² for exponential fit: %s\n", formatFloatSmart(rLn*rLn, useRounding)))
		}
	} else if regressionType == "power" {
		for i, xi := range x {
			if xi <= 0 {
				outputText.WriteString(fmt.Sprintf("\nCannot perform power regression: x value at index %d is not positive (%s).\n", i, formatFloatSmart(xi, useRounding)))
				return outputText.String(), nil
			}
		}
		for i, yi := range y {
			if yi <= 0 {
				outputText.WriteString(fmt.Sprintf("\nCannot perform power regression: y value at index %d is not positive (%s).\n", i, formatFloatSmart(yi, useRounding)))
				return outputText.String(), nil
			}
		}

		lnX := make([]float64, n)
		lnY := make([]float64, n)
		for i := 0; i < n; i++ {
			lnX[i] = math.Log(x[i])
			lnY[i] = math.Log(y[i])
		}

		sumLnX := 0.0
		sumLnY := 0.0
		sumLnXLnY := 0.0
		sumLnX2 := 0.0
		for i := 0; i < n; i++ {
			sumLnX += lnX[i]
			sumLnY += lnY[i]
			sumLnXLnY += lnX[i] * lnY[i]
			sumLnX2 += lnX[i] * lnX[i]
		}

		termLnX := float64(n)*sumLnX2 - sumLnX*sumLnX
		if termLnX == 0 {
			outputText.WriteString("\nCannot perform power regression: denominator for b is zero (all X values may be the same).\n")
			return outputText.String(), nil
		}
		b := (float64(n)*sumLnXLnY - sumLnX*sumLnY) / termLnX
		lnA := (sumLnY - b*sumLnX) / float64(n)
		a := math.Exp(lnA)

		if a == 0 || math.IsNaN(a) || math.IsInf(a, 0) {
			outputText.WriteString("\nCould not calculate power regression: invalid coefficient a (possibly due to data scale or input values).\n")
			return outputText.String(), nil
		}

		outputText.WriteString(fmt.Sprintf("\nPower equation: y = %s*x^%s\n",
			formatFloatSmart(a, useRounding), formatFloatSmart(b, useRounding)))

		if b != 0 {
			outputText.WriteString(fmt.Sprintf("Solving for x: x = (y/%s)^(1/%s)\n",
				formatFloatSmart(a, useRounding), formatFloatSmart(b, useRounding)))
		}

		sumLnY2 := 0.0
		for i := 0; i < n; i++ {
			sumLnY2 += lnY[i] * lnY[i]
		}

		numeratorLn := float64(n)*sumLnXLnY - sumLnX*sumLnY
		termLnY := float64(n)*sumLnY2 - sumLnY*sumLnY
		rDenominatorLn := termLnX * termLnY

		if rDenominatorLn > 0 {
			rLn := numeratorLn / math.Sqrt(rDenominatorLn)
			outputText.WriteString(fmt.Sprintf("r² for power fit: %s\n", formatFloatSmart(rLn*rLn, useRounding)))
		}
	} else {
		if termX != 0 {
			slope := numerator / termX
			intercept := (sumY - slope*sumX) / float64(n)
			outputText.WriteString(fmt.Sprintf("\nTrendline equation: y = %sx + %s\n",
				formatFloatSmart(slope, useRounding), formatFloatSmart(intercept, useRounding)))
		}
	}

	return outputText.String(), nil
}

type RequestPayload struct {
	XValues        string `json:"x_values"`
	YValues        string `json:"y_values"`
	UseDefault     bool   `json:"use_default"`
	XAxisLabel     string `json:"x_axis_label"`
	YAxisLabel     string `json:"y_axis_label"`
	RegressionType string `json:"regression_type"`
	UseRounding    bool   `json:"use_rounding"`
}

type ResponsePayload struct {
	TextOutput     string    `json:"text_output"`
	X              []float64 `json:"x"`
	Y              []float64 `json:"y"`
	RegressionType string    `json:"regression_type,omitempty"`
	Error          string    `json:"error,omitempty"`
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func handleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload RequestPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponsePayload{Error: "Invalid request body: " + err.Error()})
		return
	}

	var x, y []float64
	var defaultX, defaultY []float64

	switch payload.RegressionType {
	case "quadratic":
		defaultX = []float64{-3, -2, -1, 0, 1, 2, 3, 4, 5}
		defaultY = []float64{9.2, 4.1, 1.1, 0.2, 1.1, 4.2, 9.1, 16.0, 25.1}
	case "exponential":
		defaultX = []float64{0, 1, 2, 3, 4, 5}
		defaultY = []float64{2.1, 5.4, 14.8, 40.2, 109.6, 298.1}
	case "power":
		defaultX = []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		defaultY = []float64{1.1, 4.2, 9.1, 16.2, 25.1, 36.2, 49.1, 64.2, 81.1, 100.2}
	default:
		defaultX = []float64{0, 3, 5, 6, 7, 10, 12, 13, 15, 18}
		defaultY = []float64{8.2, 7.5, 7.0, 6.5, 7.2, 6.1, 6.8, 5.5, 5.8, 5.2}
	}

	w.Header().Set("Content-Type", "application/json")

	if payload.UseDefault {
		x = defaultX
		y = defaultY
	} else {
		xStrings := strings.Fields(strings.TrimSpace(payload.XValues))
		for _, sVal := range xStrings {
			val, errParse := strconv.ParseFloat(sVal, 64)
			if errParse != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ResponsePayload{Error: fmt.Sprintf("Invalid input for X values: '%s'. %v", sVal, errParse)})
				return
			}
			x = append(x, val)
		}

		yStrings := strings.Fields(strings.TrimSpace(payload.YValues))
		if len(yStrings) != len(xStrings) && len(xStrings) > 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ResponsePayload{Error: "Number of Y values must match number of X values."})
			return
		}
		for _, sVal := range yStrings {
			val, errParse := strconv.ParseFloat(sVal, 64)
			if errParse != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(ResponsePayload{Error: fmt.Sprintf("Invalid input for Y values: '%s'. %v", sVal, errParse)})
				return
			}
			y = append(y, val)
		}
		if len(x) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ResponsePayload{Error: "X values are empty and not using default."})
			return
		}
	}

	textResult, errCalc := calculateAndGenerateOutput(x, y, payload.RegressionType, payload.UseRounding)
	if errCalc != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ResponsePayload{Error: "Error during calculation: " + errCalc.Error()})
		return
	}

	response := ResponsePayload{
		TextOutput:     textResult,
		X:              x,
		Y:              y,
		RegressionType: payload.RegressionType,
	}
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/", handleForm)
	http.HandleFunc("/calculate", handleCalculate)

	http.HandleFunc("/front.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "front.js")
	})

	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "style.css")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "1337"
		fmt.Println("Warning: PORT environment variable not set, defaulting to " + port)
	}

	fmt.Println("Server starting on http://0.0.0.0:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
