package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		fmt.Println(err.Error())
	}

	s := server.NewMCPServer(
		"Calculator Demo",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	lastQuoteTool := mcp.NewTool("last_quote",
		mcp.WithDescription("Get the latest price and volume information for a ticker of your choice."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("The symbol of the global ticker of your choice. For example: symbol=IBM."),
		),
	)

	timeSeriesDailyTool := mcp.NewTool("time_series_daily",
		mcp.WithDescription("Returns raw (as-traded) daily time series (date, daily open, daily high, daily low, daily close, daily volume) of the global equity specified, covering 20+ years of historical data. The OHLCV data is sometimes called 'candles' in finance literature."),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("The symbol of the global ticker of your choice. For example: symbol=IBM.")),
		mcp.WithString("outputsize",
			mcp.Description("By default, outputsize=compact. Strings compact and full are accepted with the following specifications: compact returns only the latest 100 data points; full returns the full-length time series of 20+ years of historical data. The 'compact' option is recommended if you would like to reduce the data size of each API call."),
			mcp.Enum("compact", "full")))

	// Add the last quote handler
	s.AddTool(lastQuoteTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		symbol, err := request.RequireString("symbol")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", symbol, os.Getenv("AV_API_KEY"))

		resp, err := http.Get(url)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("unable to read request response", err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Data: %s", respBody)), nil
	})

	// Daily time series data
	s.AddTool(timeSeriesDailyTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		symbol, err := request.RequireString("symbol")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		outputsize := request.GetString("outputsize", "compact")

		url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&outputsize=%s&apikey=%s", symbol, outputsize, os.Getenv("AV_API_KEY"))

		resp, err := http.Get(url)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("unable to read request response", err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Data: %s", respBody)), nil
	})

	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
