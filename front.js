let chartInstance = null;

document.getElementById("calcForm").addEventListener("submit", handleSubmit);

async function handleSubmit(event) {
  event.preventDefault();
  const useDefault = document.getElementById("useDefault").checked;
  const regressionType = document.querySelector(
    'input[name="regressionType"]:checked'
  ).value;
  const useWholeNumberAxes =
    document.getElementById("useWholeNumberAxes").checked;
  const useRounding = document.getElementById("useRounding").checked;

  let payload = {
    regression_type: regressionType,
    use_rounding: useRounding,
  };

  if (useDefault) {
    payload.use_default = true;
  } else {
    payload.use_default = false;
    payload.x_values = document.getElementById("x_values").value;
    payload.y_values = document.getElementById("y_values").value;
  }

  document.getElementById("resultsText").innerText = "Calculating...";
  document.getElementById("plotCanvas").style.display = "none";

  try {
    const response = await fetch("/calculate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const result = await response.json();

    console.log("Response from server:", result);

    if (result.error) {
      console.error("Error:", result.error);
      document.getElementById("resultsText").innerText =
        "Error: " + result.error;
    } else {
      console.log("X values:", result.x);
      console.log("Y values:", result.y);
      document.getElementById("resultsText").innerText = result.text_output;

      if (
        result.x &&
        result.y &&
        result.x.length === result.y.length &&
        result.x.length > 0
      ) {
        const xArr = result.x.map(Number);
        const yArr = result.y.map(Number);
        const dataPoints = xArr.map((xVal, i) => ({
          x: xVal,
          y: yArr[i],
        }));
        const xMean = xArr.reduce((a, b) => a + b, 0) / xArr.length;
        const xShifted = xArr.map((x) => x - xMean);

        const n = xArr.length;
        const sumX = xArr.reduce((a, b) => a + b, 0);
        const sumY = yArr.reduce((a, b) => a + b, 0);
        const sumXY = xArr.reduce((sum, x, i) => sum + x * yArr[i], 0);
        const sumX2 = xArr.reduce((sum, x) => sum + x * x, 0);
        const sumY2 = yArr.reduce((sum, y) => sum + y * y, 0);

        let trendlinePoints = [];
        const padding = 0.05 * (Math.max(...xArr) - Math.min(...xArr) || 1);
        let axisMinX, axisMaxX, axisMinY, axisMaxY;

        if (useWholeNumberAxes) {
          axisMinX = Math.floor(Math.min(...xArr) - padding);
          axisMaxX = Math.ceil(Math.max(...xArr) + padding);
          axisMinY = Math.floor(Math.min(...yArr) - padding);
          axisMaxY = Math.ceil(Math.max(...yArr) + padding);
        } else {
          axisMinX = Math.min(...xArr) - padding;
          axisMaxX = Math.max(...xArr) + padding;
          axisMinY = Math.min(...yArr) - padding;
          axisMaxY = Math.max(...yArr) + padding;
        }

        if (regressionType === "linear") {
          // Linear regression calculation
          const slope = (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX);
          const intercept = (sumY - slope * sumX) / n;

          trendlinePoints = [
            { x: axisMinX, y: slope * axisMinX + intercept },
            { x: axisMaxX, y: slope * axisMaxX + intercept },
          ];
        } else if (regressionType === "quadratic") {
          // Quadratic regression calculation
          const sumX3 = xArr.reduce((sum, x) => sum + x * x * x, 0);
          const sumX4 = xArr.reduce((sum, x) => sum + x * x * x * x, 0);
          const sumX2Y = xArr.reduce((sum, x, i) => sum + x * x * yArr[i], 0);

          // Matrix to solve: [a, b, c] = inverse([[sumX4, sumX3, sumX2], [sumX3, sumX2, sumX], [sumX2, sumX, n]]) * [sumX2Y, sumXY, sumY]
          const det =
            sumX4 * (sumX2 * n - sumX * sumX) -
            sumX3 * (sumX3 * n - sumX * sumX2) +
            sumX2 * (sumX3 * sumX - sumX2 * sumX2);

          if (det !== 0) {
            const a =
              (sumX2Y * (sumX2 * n - sumX * sumX) -
                sumXY * (sumX3 * n - sumX * sumX2) +
                sumY * (sumX3 * sumX - sumX2 * sumX2)) /
              det;

            const b =
              (sumX4 * (sumXY * n - sumY * sumX) -
                sumX3 * (sumX2Y * n - sumY * sumX2) +
                sumX2 * (sumX2Y * sumX - sumXY * sumX2)) /
              det;

            const c =
              (sumX4 * (sumX2 * sumY - sumX * sumXY) -
                sumX3 * (sumX3 * sumY - sumX * sumX2Y) +
                sumX2 * (sumX3 * sumXY - sumX2 * sumX2Y)) /
              det;

            const step = (axisMaxX - axisMinX) / 100;
            for (let x = axisMinX; x <= axisMaxX; x += step) {
              const y = a * x * x + b * x + c;
              trendlinePoints.push({ x, y });
            }
          }
        } else if (regressionType === "exponential") {
          const logYArr = yArr.map((y) => Math.log(y));
          const sumLogY = logYArr.reduce((a, b) => a + b, 0);
          const sumXLogY = xShifted.reduce(
            (sum, x, i) => sum + x * logYArr[i],
            0
          );
          const sumX2 = xShifted.reduce((sum, x) => sum + x * x, 0);

          const b = (n * sumXLogY - 0 * sumLogY) / (n * sumX2 - 0 * 0); // sumX is zero for shifted x
          const a = Math.exp((sumLogY - b * 0) / n); // sumX is zero for shifted x

          trendlinePoints = xArr.map((xVal, i) => {
            const xS = xVal - xMean;
            return { x: xVal, y: a * Math.exp(b * xS) };
          });
        } else if (regressionType === "power") {
          // Power regression: y = a * x^b
          const logXArr = xArr.map((x) => Math.log(x));
          const logYArr = yArr.map((y) => Math.log(y));
          const sumLogX = logXArr.reduce((a, b) => a + b, 0);
          const sumLogY = logYArr.reduce((a, b) => a + b, 0);
          const sumLogXLogY = logXArr.reduce(
            (sum, logX, i) => sum + logX * logYArr[i],
            0
          );
          const sumLogX2 = logXArr.reduce((sum, logX) => sum + logX * logX, 0);

          const b =
            (n * sumLogXLogY - sumLogX * sumLogY) /
            (n * sumLogX2 - sumLogX * sumLogX);
          const a = Math.exp((sumLogY - b * sumLogX) / n);

          // Generate trendline points across the entire axis range
          const step = (axisMaxX - axisMinX) / 100;
          trendlinePoints = [];
          for (let x = axisMinX; x <= axisMaxX; x += step) {
            if (x > 0) {
              // Power function only defined for positive x
              const y = a * Math.pow(x, b);
              trendlinePoints.push({ x, y });
            }
          }
        }

        const ctx = document.getElementById("plotCanvas").getContext("2d");
        document.getElementById("plotCanvas").style.display = "block";

        if (chartInstance) {
          chartInstance.destroy();
        }

        const xLabel = document.getElementById("x_axis_label").value || "X";
        const yLabel = document.getElementById("y_axis_label").value || "Y";

        chartInstance = new Chart(ctx, {
          type: "scatter",
          data: {
            datasets: [
              {
                label: "Scatter Plot",
                data: dataPoints,
                backgroundColor: "rgba(54, 162, 235, 0.7)",
                pointRadius: 5,
              },
              {
                label:
                  regressionType === "linear"
                    ? "Linear Trendline"
                    : regressionType === "quadratic"
                    ? "Quadratic Trendline"
                    : regressionType === "exponential"
                    ? "Exponential Trendline"
                    : regressionType === "power"
                    ? "Power Trendline"
                    : "Trendline",
                data: trendlinePoints,
                type: "line",
                fill: false,
                borderColor: "rgba(255,99,132,1)",
                borderWidth: 2,
                pointRadius: 0,
                tension: regressionType === "linear" ? 0 : 0.4, // Use curve for non-linear models
              },
            ],
          },
          options: {
            devicePixelRatio: 2,
            responsive: true,
            plugins: {
              legend: { display: true },
              title: { display: true, text: `${xLabel} vs. ${yLabel}` },
            },
            scales: {
              x: {
                title: { display: true, text: xLabel },
                type: "linear",
                min: axisMinX,
                max: axisMaxX,
                ticks: {
                  stepSize: useWholeNumberAxes ? 1 : undefined,
                },
              },
              y: {
                title: { display: true, text: yLabel },
                type: "linear",
                min: axisMinY,
                max: axisMaxY,
                ticks: {
                  stepSize: useWholeNumberAxes ? 1 : undefined,
                },
              },
            },
          },
        });

        document.getElementById("exportButton").style.display = "block";
        document.getElementById("copyButton").style.display = "block";
      } else {
        document.getElementById("exportButton").style.display = "none";
        document.getElementById("copyButton").style.display = "none";
      }
    }
  } catch (e) {
    console.error("Request failed:", e);
    document.getElementById("resultsText").innerText = "Request failed: " + e;
  }
}

document.getElementById("exportButton").addEventListener("click", () => {
  const canvas = document.getElementById("plotCanvas");
  const link = document.createElement("a");
  link.download = "graph.png";
  link.href = canvas.toDataURL("image/png");
  link.click();
});

document.getElementById("copyButton").addEventListener("click", async () => {
  const canvas = document.getElementById("plotCanvas");
  try {
    const blob = await new Promise((resolve) =>
      canvas.toBlob(resolve, "image/png")
    );
    await navigator.clipboard.write([new ClipboardItem({ "image/png": blob })]);
    alert("Graph copied to clipboard!");
  } catch (error) {
    alert("Failed to copy graph to clipboard.");
  }
});
