.DEFAULT_GOAL := report

.PHONY: title report

title:
	latexmk -pdf -interaction=nonstopmode -halt-on-error title.tex

report: title
	latexmk -pdf -interaction=nonstopmode -halt-on-error report.tex
