package main

import (
	"testing"
)

func TestTemplateRendering(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     TemplateData
		expected string
		wantErr  bool
	}{
		{
			name:     "Simple repo name",
			template: "{{ .Repo.Name }}",
			data: TemplateData{
				Repo: RepoInfo{Name: "myrepo"},
			},
			expected: "myrepo",
			wantErr:  false,
		},
		{
			name:     "Repo owner",
			template: "{{ .Repo.Owner }}",
			data: TemplateData{
				Repo: RepoInfo{Owner: "admin"},
			},
			expected: "admin",
			wantErr:  false,
		},
		{
			name:     "Pipeline branch",
			template: "{{ .Pipeline.Branch }}",
			data: TemplateData{
				Pipeline: PipelineInfo{Branch: "main"},
			},
			expected: "main",
			wantErr:  false,
		},
		{
			name:     "Complex path template",
			template: "{{ .Repo.Name }}/{{ .Pipeline.Branch }}",
			data: TemplateData{
				Repo:     RepoInfo{Name: "myproject"},
				Pipeline: PipelineInfo{Branch: "develop"},
			},
			expected: "myproject/develop",
			wantErr:  false,
		},
		{
			name:     "Full name template",
			template: "{{ .Repo.FullName }}",
			data: TemplateData{
				Repo: RepoInfo{FullName: "admin/myrepo"},
			},
			expected: "admin/myrepo",
			wantErr:  false,
		},
		{
			name:     "Static value",
			template: "dronefiles",
			data:     TemplateData{},
			expected: "dronefiles",
			wantErr:  false,
		},
		{
			name:     "Mixed static and dynamic",
			template: "configs/{{ .Repo.Owner }}/{{ .Repo.Name }}",
			data: TemplateData{
				Repo: RepoInfo{Owner: "team", Name: "backend"},
			},
			expected: "configs/team/backend",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := renderTemplate(tt.template, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("renderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != tt.expected {
				t.Errorf("renderTemplate() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTemplateWithRealScenarios(t *testing.T) {
	scenarios := []struct {
		name      string
		namespace string
		reponame  string
		branch    string
		path      string
		request   ConfigRequest
		expected  map[string]string
	}{
		{
			name:      "Standard Gitea setup",
			namespace: "{{ .Repo.Owner }}",
			reponame:  "dronefiles",
			branch:    "{{ .Pipeline.Branch }}",
			path:      "{{ .Repo.Name }}/{{ .Pipeline.Branch }}",
			request: ConfigRequest{
				Repo: RepoInfo{
					Name:  "myapp",
					Owner: "admin",
				},
				Pipeline: PipelineInfo{
					Branch: "main",
				},
			},
			expected: map[string]string{
				"namespace": "admin",
				"reponame":  "dronefiles",
				"branch":    "main",
				"path":      "myapp/main",
			},
		},
		{
			name:      "Multi-tenant setup",
			namespace: "{{ .Repo.Owner }}",
			reponame:  "ci-configs",
			branch:    "master",
			path:      "pipelines/{{ .Repo.Name }}",
			request: ConfigRequest{
				Repo: RepoInfo{
					Name:  "frontend",
					Owner: "team-alpha",
				},
				Pipeline: PipelineInfo{
					Branch: "develop",
				},
			},
			expected: map[string]string{
				"namespace": "team-alpha",
				"reponame":  "ci-configs",
				"branch":    "master",
				"path":      "pipelines/frontend",
			},
		},
		{
			name:      "Branch-specific configs",
			namespace: "{{ .Repo.Owner }}",
			reponame:  "{{ .Repo.Name }}-ci",
			branch:    "{{ .Pipeline.Branch }}",
			path:      "configs/{{ .Pipeline.Branch }}",
			request: ConfigRequest{
				Repo: RepoInfo{
					Name:  "api-service",
					Owner: "backend-team",
				},
				Pipeline: PipelineInfo{
					Branch: "feature/auth",
				},
			},
			expected: map[string]string{
				"namespace": "backend-team",
				"reponame":  "api-service-ci",
				"branch":    "feature/auth",
				"path":      "configs/feature/auth",
			},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			data := TemplateData{
				Repo:     sc.request.Repo,
				Pipeline: sc.request.Pipeline,
			}

			// Test namespace
			namespace, err := renderTemplate(sc.namespace, data)
			if err != nil {
				t.Errorf("namespace template error: %v", err)
			}
			if namespace != sc.expected["namespace"] {
				t.Errorf("namespace = %q, want %q", namespace, sc.expected["namespace"])
			}

			// Test reponame
			reponame, err := renderTemplate(sc.reponame, data)
			if err != nil {
				t.Errorf("reponame template error: %v", err)
			}
			if reponame != sc.expected["reponame"] {
				t.Errorf("reponame = %q, want %q", reponame, sc.expected["reponame"])
			}

			// Test branch
			branch, err := renderTemplate(sc.branch, data)
			if err != nil {
				t.Errorf("branch template error: %v", err)
			}
			if branch != sc.expected["branch"] {
				t.Errorf("branch = %q, want %q", branch, sc.expected["branch"])
			}

			// Test path
			path, err := renderTemplate(sc.path, data)
			if err != nil {
				t.Errorf("path template error: %v", err)
			}
			if path != sc.expected["path"] {
				t.Errorf("path = %q, want %q", path, sc.expected["path"])
			}
		})
	}
}
