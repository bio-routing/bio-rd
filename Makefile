all: gazelle test coverage

gazelle:
	bazel run //:gazelle -- update

test:
	bazel test //...

coverage:
	bazel coverage //...

dep:
	bazel build //vendor/github.com/golang/dep/cmd/dep

dep-clean:
	# hack: dep of dep gives us these, and it breaks gazelle
	rm -rf vendor/github.com/golang/dep/cmd/dep/testdata
	rm -rf vendor/github.com/golang/dep/internal/fs/testdata/symlinks/dir-symlink

clean: dep-clean
	bazel clean

.PHONY: gazelle test coverage dep dep-clean clean



