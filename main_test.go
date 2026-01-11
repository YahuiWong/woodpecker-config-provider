package main

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

// Woodpecker 的 pipeline 结构
type Pipeline struct {
	Steps []Step `yaml:"steps"`
	When  []When `yaml:"when,omitempty"`
}

type Step struct {
	Name     string   `yaml:"name"`
	Image    string   `yaml:"image"`
	Commands []string `yaml:"commands"`
}

type When struct {
	Event  string `yaml:"event"`
	Branch string `yaml:"branch"`
}

func TestWoodpeckerParseConfigResponse(t *testing.T) {
	// 这是 config provider 返回的 JSON（简化版，只测试 build.yml）
	jsonResponse := `{
		"configs": [
			{
				"name": "build",
				"data": "steps:\n  - name: build\n    image: alpine\n    commands:\n      - echo \"==========================================\"\n      - echo \"Build Pipeline\"\n      - echo \"Repository: myrepo\"\n      - echo \"Branch: main\"\n      - echo \"Status: Building...\"\n      - sleep 1\n      - echo \"Build completed\"\n"
			}
		]
	}`

	t.Logf("测试 JSON: %s\n", jsonResponse)

	// 1. 解析 JSON 响应（模拟 Woodpecker）
	var configResp ConfigResponse
	err := json.Unmarshal([]byte(jsonResponse), &configResp)
	if err != nil {
		t.Fatalf("❌ JSON 解析失败: %v", err)
	}

	t.Logf("✅ JSON 解析成功，找到 %d 个配置", len(configResp.Configs))

	if len(configResp.Configs) == 0 {
		t.Fatal("❌ 没有找到配置")
	}

	config := configResp.Configs[0]
	t.Logf("配置名称: %s", config.Name)
	t.Logf("配置内容长度: %d 字节", len(config.Data))
	t.Logf("配置内容:\n%s", config.Data)

	// 2. 模拟 Woodpecker: Data: []byte(config.Data)
	yamlBytes := []byte(config.Data)

	// 3. 尝试解析为 Pipeline 结构
	var pipeline Pipeline
	err = yaml.Unmarshal(yamlBytes, &pipeline)
	if err != nil {
		t.Errorf("❌ YAML 解析失败: %v", err)

		// 尝试解析为 interface{} 看看实际结构
		var raw interface{}
		yaml.Unmarshal(yamlBytes, &raw)
		t.Logf("实际解析结果类型: %T", raw)
		t.Logf("实际解析结果: %+v", raw)
		t.FailNow()
	}

	t.Logf("✅ YAML 解析成功")
	t.Logf("Steps 数量: %d", len(pipeline.Steps))

	if len(pipeline.Steps) == 0 {
		t.Fatal("❌ 没有找到 steps")
	}

	// 验证 step 结构
	step := pipeline.Steps[0]
	t.Logf("Step 名称: %s", step.Name)
	t.Logf("Step 镜像: %s", step.Image)
	t.Logf("Step 命令数量: %d", len(step.Commands))

	// 验证命令
	expectedCommands := []string{
		`echo "=========================================="`,
		`echo "Build Pipeline"`,
		`echo "Repository: myrepo"`,
		`echo "Branch: main"`,
		`echo "Status: Building..."`,
		`sleep 1`,
		`echo "Build completed"`,
	}

	if len(step.Commands) != len(expectedCommands) {
		t.Errorf("❌ 命令数量不匹配: 期望 %d，实际 %d", len(expectedCommands), len(step.Commands))
	}

	for i, cmd := range step.Commands {
		t.Logf("  命令[%d]: %s", i, cmd)
		if i < len(expectedCommands) && cmd != expectedCommands[i] {
			t.Errorf("❌ 命令[%d]不匹配:\n  期望: %s\n  实际: %s", i, expectedCommands[i], cmd)
		}
	}
}

// 测试所有三个配置文件
func TestAllConfigs(t *testing.T) {
	// 完整的三个配置
	jsonResponse := `{
		"configs": [
			{
				"name": "build",
				"data": "steps:\n  - name: build\n    image: alpine\n    commands:\n      - echo \"==========================================\"\n      - echo \"Build Pipeline\"\n      - echo \"Repository: myrepo\"\n      - echo \"Branch: main\"\n      - echo \"Status: Building...\"\n      - sleep 1\n      - echo \"Build completed\"\n"
			},
			{
				"name": "test",
				"data": "steps:\n  - name: test\n    image: alpine\n    commands:\n      - echo \"Test Pipeline\"\n      - echo \"Test 1: Syntax check... PASS\"\n      - echo \"Test 2: Unit tests... PASS\"\n      - echo \"Result: All passed\"\n"
			},
			{
				"name": "deploy",
				"data": "when:\n  - event: push\n    branch: main\n\nsteps:\n  - name: deploy\n    image: alpine\n    commands:\n      - echo \"Deploy Pipeline\"\n      - echo \"Target: Production\"\n      - echo \"Status: Deploying...\"\n      - sleep 1\n      - echo \"Deployment successful\"\n"
			}
		]
	}`

	var configResp ConfigResponse
	err := json.Unmarshal([]byte(jsonResponse), &configResp)
	if err != nil {
		t.Fatalf("❌ JSON 解析失败: %v", err)
	}

	t.Logf("✅ 找到 %d 个配置", len(configResp.Configs))

	// 逐个测试
	for i, config := range configResp.Configs {
		t.Run(config.Name, func(t *testing.T) {
			t.Logf("测试配置 [%d]: %s", i+1, config.Name)

			yamlBytes := []byte(config.Data)
			var pipeline Pipeline
			err := yaml.Unmarshal(yamlBytes, &pipeline)
			if err != nil {
				t.Errorf("❌ YAML 解析失败: %v", err)
				t.Logf("YAML 内容:\n%s", config.Data)

				// 显示详细错误
				var raw interface{}
				yaml.Unmarshal(yamlBytes, &raw)
				t.Logf("原始解析: %+v", raw)
				return
			}

			t.Logf("✅ YAML 解析成功")
			t.Logf("Steps: %d", len(pipeline.Steps))
			if len(pipeline.When) > 0 {
				t.Logf("When 条件: %d", len(pipeline.When))
			}
		})
	}
}
