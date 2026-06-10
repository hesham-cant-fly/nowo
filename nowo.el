;;; nowo.el --- Major mode for Nowo programming language  -*- lexical-binding: t; -*-

;;; Commentary:
;; Provides syntax highlighting, indentation, and imenu for .nowo source files.

;;; Code:

(defgroup nowo nil
  "Major mode for editing Nowo source files."
  :group 'languages)

(defcustom nowo-indent-offset 2
  "Indentation offset for `nowo-mode'."
  :type 'integer
  :group 'nowo)

(defvar nowo-mode-syntax-table
  (let ((st (make-syntax-table)))
    ;; Pairs
    (modify-syntax-entry ?\( "()" st)
    (modify-syntax-entry ?\) ")(" st)
    (modify-syntax-entry ?\{ "(}" st)
    (modify-syntax-entry ?\} "){" st)
    ;; Operators (all punctuation, no comment syntax)
    (modify-syntax-entry ?+  "." st)
    (modify-syntax-entry ?-  ". 12" st)
    (modify-syntax-entry ?\n ">" st)
    (modify-syntax-entry ?*  "." st)
    (modify-syntax-entry ?/  "." st)
    (modify-syntax-entry ?=  "." st)
    (modify-syntax-entry ?:  "." st)
    (modify-syntax-entry ?,  "." st)
    (modify-syntax-entry ?\; "." st)
    (modify-syntax-entry ?.  "." st)
    (modify-syntax-entry ??  "." st)
    (modify-syntax-entry ?!  "." st)
    (modify-syntax-entry ?<  "." st)
    (modify-syntax-entry ?>  "." st)
    st)
  "Syntax table for `nowo-mode'.")

(defconst nowo-operators
  (regexp-opt '("+" "-" "*" "/" ";" "." "=" ":=" "==" "!=" "<" ">" "?" "!"))
  "Regexp matching Nowo operators.")

(defconst nowo-font-lock-keywords
  (list
   '("\\b[0-9]+\\(?:\\.[0-9]+\\)?\\b" . font-lock-constant-face)
   '("\\<\\(\\sw+\\)\\s-*:=" (1 font-lock-function-name-face))
   '("\\<\\(\\sw+\\)\\s-*(" (1 font-lock-function-name-face))
   (cons nowo-operators 'font-lock-builtin-face))
  "Font lock keywords for `nowo-mode'.")

(defun nowo-indent-line ()
  "Indent current line."
  (interactive)
  (indent-line-to
   (save-excursion
     (back-to-indentation)
     (if (memq (char-after) '(?\) ?\}))
         (ignore-errors
           (forward-char)
           (backward-sexp)
           (current-column))
       (forward-line -1)
       (while (and (not (bobp)) (looking-at "[ \t]*$"))
         (forward-line -1))
       (skip-chars-forward " \t")
       (if (eolp)
           0
         (let ((indent (current-column)))
           (end-of-line)
           (skip-chars-backward " \t")
           (pcase (char-before (point))
             ((or ?\{ ?\() (+ indent nowo-indent-offset))
             (?=
              (save-excursion
                (backward-char)
                (if (eq (char-before) ?:)
                    (+ indent nowo-indent-offset)
                  indent)))
             (_ indent))))))))

(defvar nowo-mode-map
  (let ((map (make-sparse-keymap)))
    map)
  "Keymap for `nowo-mode'.")

;;;###autoload
(define-derived-mode nowo-mode prog-mode "Nowo"
  "Major mode for editing Nowo source files.
\\{nowo-mode-map}"
  :syntax-table nowo-mode-syntax-table
  (setq font-lock-defaults '(nowo-font-lock-keywords nil nil))
  (setq-local comment-start "--")
  (setq-local comment-end "")
  (setq-local indent-line-function 'nowo-indent-line)
  (setq-local imenu-generic-expression
              (list (list nil "\\<\\(\\sw+\\)\\s-*:" 1))))

;;;###autoload
(add-to-list 'auto-mode-alist '("\\.nowo\\'" . nowo-mode))

(provide 'nowo)
;;; nowo.el ends here
