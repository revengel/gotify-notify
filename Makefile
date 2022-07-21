build:
	go mod download && \
    go build -v -o ./gotify-notify


install-script: build
	cp ./gotify-notify $(HOME)/.local/bin/gotify-notify


install-service:
	cp ./gotify-notify.service $(HOME)/.config/systemd/user/gotify-notify.service


install-all: install-script install-service




uninstall-service:
	systemctl --user stop gotify-notify.service
	systemctl --user disable gotify-notify.service
	rm $(HOME)/.config/systemd/user/gotify-notify.service


uninstall-script:
	rm $(HOME)/.local/bin/gotify-notify


uninstall-all: uninstall-service uninstall-script
