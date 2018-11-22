
;; https://www.gnu.org/software/emacs/manual/html_node/elisp/Network-Processes.html
;; (setq pserv (make-network-process :name "nu-docs" :service 3333))



(defun nudoc-send-string (c)
  (process-send-string pserv (format "%s%c%d" c last-command-event (point))))

(defun nudocs-post-command-hook ()
  (setq c "")
  (cond ((= last-command-event 127) (nudoc-send-string "d"))
        ((= last-command-event 4) (nudoc-send-string "d"))
        ((> last-command-event 64) (nudoc-send-string "i"))))

(defun handle-server-reply (process content)
  ;; get messages from server, execute commands it specifies
  ;; parse command, i/d, char, position
  (message content))

(defun nudocs-mode-enter ()
  "Called when entering nudocs mode"
  (progn
    (make-variable-buffer-local
     (defvar pserv (make-network-process :name "nudocs" :service 3333)))
    (add-hook 'post-command-hook 'nudocs-post-command-hook nil t)
    (set-process-filter pserv 'handle-server-reply)))

(defun nudocs-mode-exit ()
  (remove-hook 'post-command-hook 'nudocs-post-command-hook t))

(define-minor-mode nudocs-mode
  "Minor mode for using NUDOCS"
  :lighter " nudocs"

  (if nudocs-mode
      (nudocs-mode-enter)
    (nudocs-mode-exit)))


  ;; (progn
  ;;   (setup-nudoc-connections)
  ;;   (add-hook 'post-command-hook 'nudocs-post-command-hook nil t)
  ;;   (set-process-filter pserv 'handle-server-reply)))


;; (provide 'nudocs-mode)





;; TODO
;; get messages from server, execute commands it specifies
;; convert the code above into a minor-mode
