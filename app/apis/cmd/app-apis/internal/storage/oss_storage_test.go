package storage

import "testing"

func TestNewOSSStorageBuildsStableObjectURL(t *testing.T) {
	storage, err := NewOSSStorage(OSSConfig{
		Region:          " cn-hangzhou ",
		Endpoint:        " https://oss-cn-hangzhou.aliyuncs.com ",
		Bucket:          " resume-bucket ",
		AccessKeyID:     " access-key-id ",
		AccessKeySecret: " access-key-secret ",
		ObjectURLPrefix: " https://resume.example/base/ ",
	})
	if err != nil {
		t.Fatalf("NewOSSStorage() error = %v", err)
	}

	got := storage.URL("resumes/42/CV v1.pdf")
	want := "https://resume.example/base/resumes/42/CV%20v1.pdf"
	if got != want {
		t.Fatalf("URL() = %q, want %q", got, want)
	}
}

func TestNewOSSStorageRejectsInvalidConfig(t *testing.T) {
	valid := OSSConfig{
		Region:          "cn-hangzhou",
		Endpoint:        "https://oss-cn-hangzhou.aliyuncs.com",
		Bucket:          "resume-bucket",
		AccessKeyID:     "access-key-id",
		AccessKeySecret: "access-key-secret",
		ObjectURLPrefix: "https://resume.example",
	}
	tests := []struct {
		name   string
		mutate func(*OSSConfig)
	}{
		{name: "missing region", mutate: func(config *OSSConfig) { config.Region = "" }},
		{name: "missing access key", mutate: func(config *OSSConfig) { config.AccessKeyID = "" }},
		{name: "non HTTPS endpoint", mutate: func(config *OSSConfig) { config.Endpoint = "http://oss.example" }},
		{name: "non HTTPS URL", mutate: func(config *OSSConfig) { config.ObjectURLPrefix = "http://resume.example" }},
		{name: "URL query", mutate: func(config *OSSConfig) { config.ObjectURLPrefix += "?token=secret" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := valid
			test.mutate(&config)
			if _, err := NewOSSStorage(config); err == nil {
				t.Fatal("NewOSSStorage() error = nil, want validation error")
			}
		})
	}
}
