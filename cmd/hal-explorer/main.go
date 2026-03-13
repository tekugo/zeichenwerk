package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	. "github.com/tekugo/zeichenwerk"
)

// HAL types
type HALLink struct {
	Href      string `json:"href"`
	Templated bool   `json:"templated,omitempty"`
	Type      string `json:"type,omitempty"`
	Title     string `json:"title,omitempty"`
}

type HALResource struct {
	Links      map[string]HALLink     `json:"_links"`
	Embedded   map[string]interface{} `json:"_embedded,omitempty"`
	Properties map[string]interface{} `json:"-"`
}

type ResponseData struct {
	Status     string
	Headers    map[string][]string
	Body       string
	HAL        *HALResource
	JSONPretty string
	Error      string
}

type HALExplorer struct {
	LastResponse *ResponseData
}

func NewHALExplorer() *HALExplorer {
	return &HALExplorer{LastResponse: &ResponseData{}}
}

func (h *HALExplorer) MakeHTTPRequest(method, url, body string) (*ResponseData, error) {
	respData := &ResponseData{
		Headers: make(map[string][]string),
	}

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		respData.Error = fmt.Sprintf("Request error: %v", err)
		return respData, err
	}

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json, application/hal+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		respData.Error = fmt.Sprintf("Request failed: %v", err)
		return respData, err
	}
	defer resp.Body.Close()

	buf := make([]byte, 4096)
	var respBody []byte
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			respBody = append(respBody, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	respData.Body = string(respBody)
	respData.Status = fmt.Sprintf("%d %s", resp.StatusCode, resp.Status)

	for k, vv := range resp.Header {
		respData.Headers[k] = vv
	}

	var jsonData interface{}
	if err := json.Unmarshal(respBody, &jsonData); err == nil {
		pretty, _ := json.MarshalIndent(jsonData, "", "  ")
		respData.JSONPretty = string(pretty)

		var hal HALResource
		if err := json.Unmarshal(respBody, &hal); err == nil {
			respData.HAL = &hal
		}
	}

	h.LastResponse = respData
	return respData, nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func main() {
	explorer := NewHALExplorer()
	ui := createUI(explorer)
	ui.Run()
}

func createUI(explorer *HALExplorer) *UI {
	builder := NewBuilder(TokyoNightTheme())

	builder.Flex("main", false, "stretch", 0).
		Flex("header", true, "start", 2).
		Static("title", "HAL Explorer").
		Class("title").
		Padding(1).
		Static("subtitle", "HTTP client for HAL JSON APIs").
		End().
		Flex("content", false, "stretch", 0).
		Flex("request-panel", false, "start", 3).
		Border("", "thick").
		Padding(1).
		Flex("request-top", true, "start", 1).
		Select("method", "GET", "GET", "POST", "POST", "PUT", "PUT", "DELETE", "DELETE", "PATCH", "PATCH", "HEAD", "HEAD", "OPTIONS", "OPTIONS").
		Padding(0, 0, 0, 1).
		Editor("url-editor").
		Hint(0, -1).
		Padding(0, 0, 1, 0).
		Button("send-btn", "Send").
		Class("primary").
		End().
		Flex("headers-section", false, "start", 2).
		Static("headers-title", "Headers").
		HRule("thin").
		Flex("headers-list", false, "start", 0).
		End().
		Flex("header-add-row", true, "start", 1).
		Static("key-label", "Key:").
		Editor("header-key").Hint(20, -1).
		Editor("header-value").Hint(30, -1).
		Button("add-header-btn", "Add").
		End().
		End().
		Flex("body-section", false, "start", 3).
		Checkbox("show-body", "Request Body", false).
		Editor("body-editor").Hint(0, -1).
		End().
		End().
		Flex("response-panel", false, "stretch", 0).
		Border("", "thick").
		Padding(1).
		Flex("response-top", true, "start", 1).
		Static("response-status", "").Class("status").
		Spacer().
		Tabs("response-tabs", "Body", "Links", "Embedded", "Headers", "Properties").
		End().
		End().
		Flex("response-content", true, "stretch", 0).
		Switcher("response-switcher", false).
		Flex("body-tab", false, "stretch", 0).
		Viewport("body-viewport", "").
		Static("body-content", "").
		End().
		End().
		End().
		Flex("links-tab", false, "stretch", 0).
		Table("links-content", NewArrayTableProvider([]string{"Rel", "Href", "Templated", "Type", "Title"}, [][]string{})).
		Hint(0, -1).
		End().
		End().
		Flex("embedded-tab", false, "stretch", 0).
		Table("embedded-content", NewArrayTableProvider([]string{"Relation", "Resource"}, [][]string{})).
		Hint(0, -1).
		End().
		End().
		Flex("headers-tab", false, "stretch", 0).
		Table("headers-content", NewArrayTableProvider([]string{"Header", "Value"}, [][]string{})).
		Hint(0, -1).
		End().
		End().
		Flex("properties-tab", false, "stretch", 0).
		Table("properties-content", NewArrayTableProvider([]string{"Property", "Value"}, [][]string{})).
		Hint(0, -1).
		End().
		End().
		End().
		End().
		End().
		End().
		Flex("footer", true, "start", 1).
		Static("status-bar", "Ready").
		Class("status").
		Spacer().
		Static("hints", "Tab: navigate • Enter: select • Ctrl+S: send • Ctrl+Q: quit").
		End().
		End()

	ui := builder.Build()

	// Set initial values
	if ed := Find(ui, "url-editor"); ed != nil {
		if editor, ok := ed.(*Editor); ok {
			editor.Load("https://api.github.com/")
		}
	}
	if ed := Find(ui, "body-editor"); ed != nil {
		if editor, ok := ed.(*Editor); ok {
			editor.Load("{\n  \"key\": \"value\"\n}")
		}
	}
	// Hide body editor initially
	if ed := Find(ui, "body-editor"); ed != nil {
		ed.SetFlag("hidden", true)
	}

	explorer.setupEventHandlers(ui)
	return ui
}

func (h *HALExplorer) updateResponseDisplay(ui *UI) {
	resp := h.LastResponse

	if resp.Error != "" {
		Update(ui, "status-bar", "Error: "+resp.Error)
	} else {
		Update(ui, "status-bar", resp.Status)
	}

	if resp.JSONPretty != "" {
		Update(ui, "body-content", []string{resp.JSONPretty})
	} else if resp.Body != "" {
		Update(ui, "body-content", []string{resp.Body})
	}

	if resp.HAL != nil && len(resp.HAL.Links) > 0 {
		linksData := make([][]string, 0, len(resp.HAL.Links))
		for rel, link := range resp.HAL.Links {
			linksData = append(linksData, []string{
				rel,
				truncateString(link.Href, 60),
				fmt.Sprintf("%v", link.Templated),
				link.Type,
				link.Title,
			})
		}
		if table := Find(ui, "links-content"); table != nil {
			if t, ok := table.(*Table); ok {
				t.Set(NewArrayTableProvider([]string{"Rel", "Href", "Templated", "Type", "Title"}, linksData))
			}
		}
	}

	if resp.HAL != nil && len(resp.HAL.Embedded) > 0 {
		embeddedData := make([][]string, 0, len(resp.HAL.Embedded))
		for rel, embedded := range resp.HAL.Embedded {
			jsonBytes, _ := json.MarshalIndent(embedded, "", "  ")
			embeddedData = append(embeddedData, []string{
				rel,
				truncateString(string(jsonBytes), 80),
			})
		}
		if table := Find(ui, "embedded-content"); table != nil {
			if t, ok := table.(*Table); ok {
				t.Set(NewArrayTableProvider([]string{"Relation", "Resource"}, embeddedData))
			}
		}
	}

	if len(resp.Headers) > 0 {
		headersData := make([][]string, 0, len(resp.Headers))
		for k, vv := range resp.Headers {
			for _, v := range vv {
				headersData = append(headersData, []string{k, v})
			}
		}
		if table := Find(ui, "headers-content"); table != nil {
			if t, ok := table.(*Table); ok {
				t.Set(NewArrayTableProvider([]string{"Header", "Value"}, headersData))
			}
		}
	}

	if resp.HAL != nil && len(resp.HAL.Properties) > 0 {
		propsData := make([][]string, 0, len(resp.HAL.Properties))
		for k, v := range resp.HAL.Properties {
			jsonBytes, _ := json.MarshalIndent(v, "", "  ")
			propsData = append(propsData, []string{
				k,
				truncateString(string(jsonBytes), 80),
			})
		}
		if table := Find(ui, "properties-content"); table != nil {
			if t, ok := table.(*Table); ok {
				t.Set(NewArrayTableProvider([]string{"Property", "Value"}, propsData))
			}
		}
	}
}

func (h *HALExplorer) setupEventHandlers(ui *UI) {
	sendBtn := Find(ui, "send-btn")
	if sendBtn != nil {
		sendBtn.On("click", func(widget Widget, event string, data ...any) bool {
			h.handleSendRequest(ui)
			return true
		})
	}

	showBodyCheckbox := Find(ui, "show-body")
	if showBodyCheckbox != nil {
		showBodyCheckbox.On("change", func(widget Widget, event string, data ...any) bool {
			if checked, ok := data[0].(bool); ok {
				editor := Find(ui, "body-editor")
				if editor != nil {
					if checked {
						editor.SetFlag("hidden", false)
					} else {
						editor.SetFlag("hidden", true)
					}
					ui.Redraw(editor)
				}
			}
			return true
		})
	}

	addHeaderBtn := Find(ui, "add-header-btn")
	if addHeaderBtn != nil {
		addHeaderBtn.On("click", func(widget Widget, event string, data ...any) bool {
			Update(ui, "status-bar", "Dynamic headers not implemented in demo")
			return true
		})
	}

	ui.On("key", func(widget Widget, event string, data ...any) bool {
		if key, ok := data[0].(string); ok {
			if key == "ctrl s" {
				h.handleSendRequest(ui)
				return true
			}
			if key == "ctrl q" {
				ui.Close()
				return true
			}
		}
		return false
	})

	tabs := Find(ui, "response-tabs")
	if tabs != nil {
		if t, ok := tabs.(*Tabs); ok {
			t.On("activate", func(widget Widget, event string, data ...any) bool {
				if tabIndex, ok := data[0].(int); ok {
					h.showResponseTab(ui, tabIndex)
				}
				return true
			})
		}
	}
}

func (h *HALExplorer) handleSendRequest(ui *UI) {
	methodWidget := Find(ui, "method")
	method := "GET"
	if s, ok := methodWidget.(*Select); ok {
		method = s.Value()
	}

	urlWidget := Find(ui, "url-editor")
	var url string
	if ed, ok := urlWidget.(*Editor); ok {
		if len(ed.Lines()) > 0 {
			url = strings.TrimSpace(ed.Lines()[0])
		}
	}
	if url == "" {
		Update(ui, "status-bar", "Please enter a URL")
		return
	}

	body := ""
	showBody := false
	if cb := Find(ui, "show-body"); cb != nil {
		if c, ok := cb.(*Checkbox); ok {
			showBody = c.Flag("checked")
		}
	}
	if showBody {
		bodyWidget := Find(ui, "body-editor")
		if bodyWidget != nil {
			if ed, ok := bodyWidget.(*Editor); ok {
				if len(ed.Lines()) > 0 {
					body = ed.Lines()[0]
				}
			}
		}
	}

	Update(ui, "status-bar", fmt.Sprintf("Sending %s to %s...", method, url))
	ui.Refresh()

	resp, err := h.MakeHTTPRequest(method, url, body)
	if err != nil {
		Update(ui, "status-bar", "Request failed: "+err.Error())
		return
	}

	h.LastResponse = resp
	h.updateResponseDisplay(ui)
	Update(ui, "status-bar", fmt.Sprintf("Received %s", resp.Status))
}

func (h *HALExplorer) showResponseTab(ui *UI, tabIndex int) {
	switcher := Find(ui, "response-switcher")
	if s, ok := switcher.(*Switcher); ok {
		s.Select(tabIndex)
	}
}
