

(defun nudocs-send-raw (pserv s)
;;  (message s)
  (process-send-string pserv s))

(defun nudocs-send-operation (pserv c p)
  (nudocs-send-raw pserv (format "%s%c%d\n" c last-command-event p)))

(defun nudocs-post-command-hook (pserv)
  (cond ((= last-command-event 127) (nudocs-send-operation pserv "d" (- (point) 1)))
        ((= last-command-event 4) (nudocs-send-operation pserv "d" (- (point) 1)))
        ((> last-command-event 64) (nudocs-send-operation pserv "i" (- (point) 2)))))

(defun nudocs-set-buffer (content)
  (progn
    (setq curr-point (point))
    ;; clear old content from buffer
    (erase-buffer)

    ;; write new doc to buffer
    (insert content)
    (goto-char curr-point)))

;; returns a list of valid doc strings
(defun nudocs-split-string (original accumulator)
  (progn
    (setq lengthContent (car (split-string original ":" t)))
    (if lengthContent
        accumulator
      (progn
        (setq length (string-to-number (car l)))

        (setq accumulator (add-to-list accumulator (substring (cdr l) 0 length)))

        ;; need to remove
        (nudocs-split-string (substring (cdr l) length nil))))))

(defun handle-server-reply (process content)
  ;; split content at demarcation character
  (mapc '(nudocs-set-buffer content) (nudocs-split-string content '())))

(defun nudocs-mode-enter ()
  "Called when entering nudocs mode"
  (progn
    (message "Starting nudocs")
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
