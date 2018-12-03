(defun nudocs-send-raw (pserv s)
  (process-send-string pserv s))

(defun nudocs-send-operation (pserv c p)
  (cond ((= last-command-event 13) (nudocs-send-raw pserv (format "%s\n%d\n" c p)))
        ((= last-command-event 127) (nudocs-send-raw pserv (format "%s0%d\n" c p)))
        (t (nudocs-send-raw pserv (format "%s%c%d\n" c last-command-event p)))))

(defun nudocs-post-command-hook (pserv)
  (print last-command-event)
  (cond ((= last-command-event 127) (nudocs-send-operation pserv "d" (- (point) 1)))
        ((= last-command-event 4) (nudocs-send-operation pserv "d" (- (point) 1)))
        ((> last-command-event 31) (nudocs-send-operation pserv "i" (- (point) 2)))
        ((= last-command-event 13) (nudocs-send-operation pserv "i" (- (point) 1)))))

(defun nudocs-set-buffer (content)
  (if (not content)
      (erase-buffer)
  (progn
    (setq curr-point (point))
    ;; clear old content from buffer
    (erase-buffer)

    ;; write new doc to buffer
    (insert content)
    (goto-char curr-point))))

;; returns a list of valid doc strings
(defun nudocs-split-string (original accumulator)
  ;; stop recursing when original is nil
  (if (not (string= "" original))
      (let* ((l (split-string original ":"))
             (length (string-to-number (car l)))
             (rest (mapconcat 'identity (cdr l) ":"))
             (doc-string (substring rest 0 length))
             (next (substring rest length nil)))
        (nudocs-split-string next (cons doc-string accumulator)))
    (reverse accumulator)))

(defun handle-server-reply (process content)
  (mapc 'nudocs-set-buffer (nudocs-split-string content '())))

(defun nudocs-mode-enter ()
  "Called when entering nudocs mode"
  (progn
    (message "Starting nudocs...")
    (shell-command "nudocs -p 3333 -h ~/.nudocs/hostsfile.txt &>~/.nudocs/nudocs.log &" 1 nil)

    ;; wait for nudocs to startup
    (sleep-for 1)

    ;; https://www.gnu.org/software/emacs/manual/html_node/elisp/Network-Processes.html
    (setq pserv (make-network-process :name "nudocs" :service 3333))
    (set-process-filter pserv 'handle-server-reply)
    (nudocs-send-raw pserv "client")
    (add-hook 'post-command-hook (lambda () (nudocs-post-command-hook pserv)) nil t)))

(defun nudocs-mode-exit ()
  (progn
    (shell-command "ps -ax | grep nudocs | grep -v grep | awk '{print $1}' | xargs kill -9" nil nil)
    (remove-hook 'post-command-hook 'nudocs-post-command-hook t)))

(define-minor-mode nudocs-mode
  "Minor mode for using NUDOCS"
  :lighter " nudocs"

  (if nudocs-mode
      (nudocs-mode-enter)
    (nudocs-mode-exit)))


(provide 'nudocs-mode)
