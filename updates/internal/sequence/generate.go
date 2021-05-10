package sequence

//go:generate go run -modfile=../../../_tools/go.mod github.com/golang/mock/mockgen -source=hooks.go -package=sequence -mock_names=hooks=MockedHooks -destination=hooks_mock_test.go
