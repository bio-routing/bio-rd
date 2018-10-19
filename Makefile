NAME=bio-rd

all: test

$(NAME): gazelle
	bazel build //:bio-rd

gazelle:
	bazel run //:gazelle -- update

test: $(NAME)
	bazel test //...

dep:
	bazel build //vendor/github.com/golang/dep/cmd/dep

dep-clean:
	# hack: dep of dep gives us these, and it breaks gazelle
	rm -rf vendor/github.com/golang/dep/cmd/dep/testdata
	rm -rf vendor/github.com/golang/dep/internal/fs/testdata/symlinks/dir-symlink

clean: dep-clean
	bazel clean
	rm $(NAME)

.PHONY: $(NAME) gazelle clean dep dep-clean
