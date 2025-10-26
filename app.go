package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectImage() (string, error) {
	return wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Select Image",
		Filters: []wailsruntime.FileFilter{
			{
				Pattern:     "*.png;*.jpg;*.jpeg;*.bmp;*.webp",
				DisplayName: "Images",
			},
		},
	})
}

func (a *App) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

type RequestData struct {
	Method string `json:"method"`
	Url    string `json:"url"`
	Data   string `json:"data"`
	Cookie string `json:"cookie"`
}

type ResponseData struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	Data    []byte              `json:"data"`
}

func (a *App) Request(payload RequestData) (*ResponseData, error) {
	client := &http.Client{}
	req, err := http.NewRequest(payload.Method, payload.Url, strings.NewReader(payload.Data))
	if err != nil {
		return &ResponseData{Status: 0, Headers: nil, Data: []byte(`{"error": "` + err.Error() + `"}`)}, nil
	}

	if payload.Cookie != "" {
		req.Header.Add("Cookie", payload.Cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return &ResponseData{Status: 0, Headers: nil, Data: []byte(`{"error": "` + err.Error() + `"}`)}, nil
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ResponseData{Status: 0, Headers: nil, Data: []byte(`{"error": "` + err.Error() + `"}`)}, nil
	}

	headers := make(map[string][]string)
	for key, values := range resp.Header {
		headers[key] = values
	}

	return &ResponseData{
		Status:  resp.StatusCode,
		Headers: headers,
		Data:    data,
	}, nil
}

func (a *App) WriteSettings(settings string) error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	exeDir := filepath.Dir(exePath)
	path := filepath.Join(exeDir, "settings.json")
	return os.WriteFile(path, []byte(settings), 0644)
}

func (a *App) ReadSettings() (*string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	exeDir := filepath.Dir(exePath)
	path := filepath.Join(exeDir, "settings.json")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	strData := string(data)
	return &strData, nil
}

func (a *App) OpenURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
