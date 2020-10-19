GH_PAGES_SOURCES = src/esa/Infra/rest/generated/html/index.html Makefile

.PHONY: gh-pages
gh-pages:
	git checkout gh-pages
	rm -rf bin pkg githooks scripts Jenkinsfile, README.md
	git checkout master $(GH_PAGES_SOURCES)
	git reset HEAD
	mv -fv src/esa/infra/rest/generated/html/index.html ./
	rm -rf $(GH_PAGES_SOURCES) build src
	git add -A
	git commit -m "Generated gh-pages for `git log master -1 --pretty=short --abbrev-commit`" && git push origin gh-pages ; git checkout master