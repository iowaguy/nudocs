

(defun nudocs-send-raw (pserv s)
  (process-send-string pserv s))

(defun nudocs-send-operation (pserv c)
  (nudocs-send-raw pserv (format "%s%c%d\n" c last-command-event (- (point) 2))))

(defun nudocs-post-command-hook (pserv)
  (cond ((= last-command-event 127) (nudocs-send-operation pserv "d"))
        ((= last-command-event 4) (nudocs-send-operation pserv "d"))
        ((> last-command-event 64) (nudocs-send-operation pserv "i"))))

(defun handle-server-reply (process content)
  ;; get messages from server, execute commands it specifies
  ;; parse command, i/d, char, position
  (message content))

(defun nudocs-mode-enter ()
  "Called when entering nudocs mode"
  (progn
    (message "Starting nudocs")
    (shell-command "nudocs -p 3333 -h ~/.nudocs/hostsfile.txt &>~/.nudocs/nudocs.log &" 1 nil)

    ;; wait for nudocs to startup
    (sleep-for 1)

    ;; https://www.gnu.org/software/emacs/manual/html_node/elisp/Network-Processes.html
    (setq pserv (make-network-process :name "nudocs" :service 3333))
    (set (make-local-variable 'peer-server) pserv)
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
