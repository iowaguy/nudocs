;; https://www.gnu.org/software/emacs/manual/html_node/elisp/Network-Processes.html
(setq pserv (make-network-process :name "hello" :service 3333))
(defun mymode-post-command-hook ()
  (setq c "")
  (cond ((= last-command-event 127) (setq c "d"))
        ((= last-command-event 4) (setq c "d"))
        ((> last-command-event 64) (setq c "i")))
  (process-send-string pserv (format "%s%c%d" c last-command-event (point))))

(add-hook 'post-command-hook 'mymode-post-command-hook nil t)


(defun handle-server-reply (process content)
   (message content))

(set-process-filter pserv 'handle-server-reply)


;; TODO
;; get messages from server, execute commands it specifies
;; convert the code above into a minor-mode
