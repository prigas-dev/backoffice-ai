package operations_test

import (
	"math"
	"testing"

	"github.com/dop251/goja"
	"github.com/prigas-dev/backoffice-ai/operations"
	"github.com/stretchr/testify/assert"
)

func TestGoja(t *testing.T) {
	t.Parallel()

	t.Run("test it works with promises", func(t *testing.T) {
		t.Parallel()

		vm := goja.New()
		defer vm.Interrupt("halt")

		v, err := vm.RunString("new Promise(function (resolve) { resolve(); })")
		assert.NoError(t, err)

		shouldBeAPromise := v.Export()
		_, isAPromise := shouldBeAPromise.(*goja.Promise)
		assert.True(t, isAPromise)
	})

	t.Run("test it resolve value", func(t *testing.T) {
		t.Parallel()

		vm := goja.New()
		defer vm.Interrupt("halt")

		v, err := vm.RunString("new Promise(function (resolve) { resolve('banana'); })")
		assert.NoError(t, err)

		promise := v.Export().(*goja.Promise)

		resolvedValue := promise.Result()

		shouldBeAString := resolvedValue.Export()

		assert.Equal(t, "banana", shouldBeAString)
	})

	t.Run("test reject value", func(t *testing.T) {
		t.Parallel()

		vm := goja.New()
		defer vm.Interrupt("halt")

		v, err := vm.RunString("class NewError extends Error {} (async function run() { throw new NewError('banana'); })()")
		assert.NoError(t, err)

		promise := v.Export().(*goja.Promise)

		rejectedValue := promise.Result()

		v, err = vm.RunString("Error")
		assert.NoError(t, err)

		errorClass := v.ToObject(vm)

		isError := vm.InstanceOf(rejectedValue, errorClass)
		assert.True(t, isError)

		errorObject := rejectedValue.ToObject(vm)

		message, messageIsString := errorObject.Get("message").Export().(string)
		stack, stackIsString := errorObject.Get("stack").Export().(string)

		assert.True(t, messageIsString)
		assert.True(t, stackIsString)
		assert.Equal(t, message, "banana")
		assert.NotEmpty(t, stack)
	})

	t.Run("template strings", func(t *testing.T) {
		t.Parallel()

		vm := goja.New()
		defer vm.Interrupt("halt")

		v, err := vm.RunString("`asdfads ${123} asdfasdf`")
		assert.NoError(t, err)

		assert.Equal(t, "asdfads 123 asdfasdf", v.Export())

	})
}

func TestExecuteJavascript(t *testing.T) {
	t.Parallel()

	t.Run("should extract object", func(t *testing.T) {
		t.Parallel()

		type Prigas struct {
			Name   string `json:"name"`
			Prigas bool   `json:"prigas"`
		}

		prigas, err := operations.ExecuteJavascript[Prigas]("",
			`function run() {
				return { name: 'prigas', prigas: true } 
			}`, map[string]any{}, map[string]any{})

		assert.NoError(t, err)
		assert.Equal(t, Prigas{Name: "prigas", Prigas: true}, prigas)
	})

	t.Run("should extract string", func(t *testing.T) {
		t.Parallel()

		str, err := operations.ExecuteJavascript[string]("",
			`function run() {
				return 'prigas'
			}`, map[string]any{}, map[string]any{})

		assert.NoError(t, err)
		assert.Equal(t, "prigas", str)
	})

	t.Run("should extract boolean", func(t *testing.T) {
		t.Parallel()

		boolean, err := operations.ExecuteJavascript[bool]("",
			`function run() {
				return true
			}`, map[string]any{}, map[string]any{})

		assert.NoError(t, err)
		assert.Equal(t, true, boolean)
	})

	t.Run("should extract integers", func(t *testing.T) {
		t.Parallel()

		integer, err := operations.ExecuteJavascript[int32]("",
			`function run() {
				return 32
			}`, map[string]any{}, map[string]any{})

		assert.NoError(t, err)
		assert.Equal(t, int32(32), integer)
	})

	t.Run("should extract floats", func(t *testing.T) {
		t.Parallel()

		floater, err := operations.ExecuteJavascript[float64]("",
			`function run() {
				return Infinity
			}`, map[string]any{}, map[string]any{})

		assert.NoError(t, err)
		assert.True(t, math.IsInf(floater, 1))
	})

	t.Run("should extract null from null", func(t *testing.T) {
		t.Parallel()

		nullable, err := operations.ExecuteJavascript[*int]("",
			`function run() {
				return null
			}`, map[string]any{}, map[string]any{})

		assert.NoError(t, err)
		assert.Nil(t, nullable)
	})

	t.Run("should extract null from undefined", func(t *testing.T) {
		t.Parallel()

		undefinable, err := operations.ExecuteJavascript[*int]("",
			`function run() {
				return undefined
			}`, map[string]any{}, map[string]any{})

		assert.NoError(t, err)
		assert.Nil(t, undefinable)
	})

	t.Run("should extract Error", func(t *testing.T) {
		t.Parallel()

		_, err := operations.ExecuteJavascript[any]("",
			`function run() {
				throw new Error('banana')
			}`, map[string]any{}, map[string]any{})

		assert.ErrorContains(t, err, "Error: banana")
	})

	t.Run("should resolve Promise", func(t *testing.T) {
		t.Parallel()

		result, err := operations.ExecuteJavascript[string]("",
			`function run() {
				return new Promise((resolve) => {
					resolve('prigas')
				})
			}`, map[string]any{}, map[string]any{})

		assert.NoError(t, err)
		assert.Equal(t, "prigas", result)
	})

	t.Run("should reject Promise", func(t *testing.T) {
		t.Parallel()

		_, err := operations.ExecuteJavascript[any]("",
			`function run() {
				return new Promise((_, reject) => {
					reject(new Error('my error'))
				})
			}`, map[string]any{}, map[string]any{})

		assert.ErrorContains(t, err, "my error")
	})

	t.Run("should reject non error values", func(t *testing.T) {
		t.Parallel()

		_, err := operations.ExecuteJavascript[any]("",
			`function run() {
				return new Promise((_, reject) => {
					reject('non error')
				})
			}`, map[string]any{}, map[string]any{})

		assert.ErrorContains(t, err, "non error")
	})

	t.Run("should run global function", func(t *testing.T) {
		t.Parallel()

		globals := map[string]any{
			"banana": func(value string) string {
				return "a" + value
			},
		}

		value, err := operations.ExecuteJavascript[any]("",
			`function run() {
				return banana('b')
			}`, map[string]any{}, globals)
		assert.NoError(t, err)

		assert.Equal(t, "ab", value)
	})
}
