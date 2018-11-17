
;;;###autoload
(define-minor-mode nu-docs-mode
  :lighter "NUDOCS"

  ;; https://www.gnu.org/software/emacs/manual/html_node/elisp/Network-Processes.html
  (setq pserv (make-network-process :name "nu-docs" :service 3333))

  (defun nudocs-post-command-hook ()
    (setq c "")
    (cond ((= last-command-event 127) (setq c "d"))
          ((= last-command-event 4) (setq c "d"))
          ((> last-command-event 64) (setq c "i")))
    (process-send-string pserv (format "%s%c%d" c last-command-event (point))))

  (defun handle-server-reply (process content)
    ;; get messages from server, execute commands it specifies

    ;; parse command, i/d, char, position

    (message content))

  (add-hook 'post-command-hook 'nudocs-post-command-hook nil t)
  (set-process-filter pserv 'handle-server-reply))


(provide 'nu-docs-mode)





;; TODO
;; get messages from server, execute commands it specifies
;; convert the code above into a minor-mode
