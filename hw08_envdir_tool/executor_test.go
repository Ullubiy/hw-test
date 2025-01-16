package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCmd(t *testing.T) {
	t.Run("Успешное выполнение команды", func(t *testing.T) {
		env := map[string]string{
			"STR1": "foo",
			"STR2": "bar",
		}
		cmd := []string{"bash", "-c", "echo $STR1$STR2"}
		code := RunCmd(cmd, env)

		fmt.Println(cmd)
		assert.Equal(t, 0, code)
	})

	t.Run("Ошибка при выполнении невалидной команды", func(t *testing.T) {
		env := map[string]string{}
		cmd := []string{"bash", "-c", "ls /xxx"}
		code := RunCmd(cmd, env)

		fmt.Println(cmd)
		assert.Equal(t, 1, code)
	})
}
