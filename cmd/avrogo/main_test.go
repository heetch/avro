package main

import (
	"github.com/actgardner/gogen-avro/v10/schema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getPackageName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		namespace string
		want      string
	}{
		{
			name:      "empty",
			namespace: "",
			want:      "",
		},
		{
			name:      "ok",
			namespace: "com.heetch.location",
			want:      "location",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := getPackageName(schema.QualifiedName{Namespace: tt.namespace})
			assert.Equal(t, tt.want, actual)
		})
	}
}
