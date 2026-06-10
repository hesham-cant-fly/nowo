;;; nowo.el --- Major mode for Nowo programming language  -*- lexical-binding: t; -*-

;;; Commentary:
;; Provides syntax highlighting for .nowo source files.

;;; Code:

(defvar nowo-mode-syntax-table
  (let ((st (make-syntax-table)))
    (modify-syntax-entry ?\( "()" st)
    (modify-syntax-entry ?\) ")(" st)
    (modify-syntax-entry ?\{ "(}" st)
    (modify-syntax-entry ?\} "){" st)
    (modify-syntax-entry ?+  "." st)
    (modify-syntax-entry ?-  "." st)
    (modify-syntax-entry ?=  "." st)
    (modify-syntax-entry ?:  "." st)
    (modify-syntax-entry ?,  "." st)
    (modify-syntax-entry ?\; "." st)
    (modify-syntax-entry ?.  "." st)
    st)
  "Syntax table for `nowo-mode'.")

(defconst nowo-font-lock-keywords
  (list
   '("--.*" . font-lock-comment-face)
   (cons "\\b[0-9]+\\(\\.[0-9]+\\)?\\b" 'font-lock-constant-face)
   (cons ":" 'font-lock-builtin-face)
   '("\\<\\(\\sw+\\)\\s-*:" (1 font-lock-function-name-face))
   '("\\<\\(\\sw+\\)\\s-*(" (1 font-lock-function-name-face)))
  "Font lock keywords for `nowo-mode'.")

;;;###autoload
(define-derived-mode nowo-mode prog-mode "Nowo"
  "Major mode for editing Nowo source files."
  (set-syntax-table nowo-mode-syntax-table)
  (setq font-lock-defaults '(nowo-font-lock-keywords))
  (setq-local comment-start "--")
  (setq-local comment-end ""))

;;;###autoload
(add-to-list 'auto-mode-alist '("\\.nowo\\'" . nowo-mode))

(provide 'nowo)
;;; nowo.el ends here
