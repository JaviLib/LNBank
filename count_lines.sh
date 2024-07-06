find . -type f -not -path '*/\.git/*' -not -name 'LICENSE' -exec file --mime {} + | grep text/ | cut -d':' -f1 | xargs wc -l
