package main

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLDirectParsing(t *testing.T) {
	// 直接测试 YAML 字符串（不经过 JSON）
	yamlStr := `steps:
  - name: build
    image: alpine
    commands:
      - echo "=========================================="
      - echo "Build Pipeline"
      - echo "Repository: myrepo"
      - echo "Branch: main"
      - echo "Status: Building..."
      - sleep 1
      - echo "Build completed"
`

	t.Log("测试直接解析 YAML 字符串")
	t.Logf("YAML 内容:\n%s", yamlStr)

	var pipeline Pipeline
	err := yaml.Unmarshal([]byte(yamlStr), &pipeline)
	if err != nil {
		t.Fatalf("❌ YAML 解析失败: %v", err)
	}

	t.Log("✅ YAML 解析成功")
	t.Logf("Commands: %v", pipeline.Steps[0].Commands)
}

func TestJSONStringEscaping(t *testing.T) {
	// 测试 JSON 字符串中的转义
	jsonStr := `{"data": "steps:\n  - name: build\n    commands:\n      - echo \"Repository: myrepo\"\n"}`

	t.Logf("JSON 字符串:\n%s", jsonStr)

	var obj struct {
		Data string `json:"data"`
	}

	err := json.Unmarshal([]byte(jsonStr), &obj)
	if err != nil {
		t.Fatalf("❌ JSON 解析失败: %v", err)
	}

	t.Log("✅ JSON 解析成功")
	t.Logf("解码后的字符串:\n%s", obj.Data)
	t.Logf("字符串字节: %v", []byte(obj.Data))

	// 打印每个字符的十六进制
	t.Log("\n字符详情:")
	for i, ch := range []byte(obj.Data) {
		t.Logf("  [%d] 0x%02x %q", i, ch, string(ch))
	}
}

func TestActualConfigProviderResponse(t *testing.T) {
	// 测试实际的 config provider 响应格式
	jsonResponse := []byte(`{"configs":[{"name":"build","data":"steps:\n  - name: build\n    image: alpine\n    commands:\n      - echo \"==========================================\"\n      - echo \"Build Pipeline\"\n      - echo \"Repository: myrepo\"\n      - echo \"Branch: main\"\n      - echo \"Status: Building...\"\n      - sleep 1\n      - echo \"Build completed\"\n"}]}`)

	t.Logf("JSON 响应 (%d 字节):\n%s\n", len(jsonResponse), string(jsonResponse))

	var resp ConfigResponse
	err := json.Unmarshal(jsonResponse, &resp)
	if err != nil {
		t.Fatalf("❌ JSON 解析失败: %v", err)
	}

	data := resp.Configs[0].Data
	t.Logf("\n解码后的 Data 字段:\n%s", data)

	// 检查字符串中是否有特殊字符
	t.Log("\nData 字段字节分析:")
	for i, b := range []byte(data) {
		if b == '"' || b == ':' || b == '\\' {
			t.Logf("  [%d] 0x%02x %q <-- 特殊字符", i, b, string(b))
		}
	}

	// 尝试解析 YAML
	t.Log("\n尝试解析为 YAML:")
	var pipeline Pipeline
	err = yaml.Unmarshal([]byte(data), &pipeline)
	if err != nil {
		t.Logf("❌ YAML 解析失败: %v", err)

		// 尝试解析为原始对象
		var raw interface{}
		yaml.Unmarshal([]byte(data), &raw)
		t.Logf("原始解析结果: %+v", raw)

		// 使用 json 格式化输出
		jsonRaw, _ := json.MarshalIndent(raw, "", "  ")
		t.Logf("原始解析结果 (JSON 格式):\n%s", string(jsonRaw))
	} else {
		t.Log("✅ YAML 解析成功")
		t.Logf("Commands: %v", pipeline.Steps[0].Commands)
	}
}

func TestMinimalCase(t *testing.T) {
	// 最小化测试用例
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "简单命令",
			yaml: `commands:
  - echo hello`,
		},
		{
			name: "带双引号的命令",
			yaml: `commands:
  - echo "hello"`,
		},
		{
			name: "带冒号的命令（无引号）",
			yaml: `commands:
  - echo hello: world`,
		},
		{
			name: "带冒号的命令（有引号）",
			yaml: `commands:
  - echo "hello: world"`,
		},
		{
			name: "实际的 echo 命令",
			yaml: `commands:
  - echo "Repository: myrepo"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("YAML:\n%s", tt.yaml)

			var result struct {
				Commands []string `yaml:"commands"`
			}

			err := yaml.Unmarshal([]byte(tt.yaml), &result)
			if err != nil {
				t.Errorf("❌ 解析失败: %v", err)

				var raw interface{}
				yaml.Unmarshal([]byte(tt.yaml), &raw)
				t.Logf("原始解析: %+v", raw)
			} else {
				t.Logf("✅ 解析成功: %v", result.Commands)
			}
		})
	}
}
