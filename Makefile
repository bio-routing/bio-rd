NAME=bio-rd

all: test

$(NAME): gazelle
	bazel build //:bio-rd

gazelle:
	bazel run //:gazelle -- update

test: $(NAME)
	bazel test //...

vendor:
	bazel build //vendor/github.com/golang/dep/cmd/dep
	bazel-bin/vendor/github.com/golang/dep/cmd/dep/linux_amd64_stripped/dep use
	# hack: dep of dep gives us these, and it breaks gazelle
	rm -rf vendor/github.com/golang/dep/cmd/dep/testdata
	rm -rf vendor/github.com/golang/dep/internal/fs/testdata/symlinks/dir-symlink

clean:
	bazel clean
	rm $(NAME)

.PHONY: $(NAME) gazelle clean
