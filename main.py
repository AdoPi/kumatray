import rumps
import requests
import threading
import time

class KumaTray(rumps.App):
    def __init__(self):
        super(KumaTray, self).__init__("KumaTray", quit_button=None)
        self.icon = None
        self.menu = ["Forcer vérification", "Quitter"]
        self.status_ok = "🟢"
        self.status_problem = "🔴"
        self.status_unknown = "🟡"
        self.api_url = "https://uptime.unova.fr/api/status"
        self.interval = 30  # secondes
        self.title = self.status_unknown
        threading.Thread(target=self.poll_loop, daemon=True).start()

    def poll_loop(self):
        while True:
            self.check_status()
            time.sleep(self.interval)

    def check_status(self):
        try:
            r = requests.get(self.api_url, timeout=5)
            txt = r.text.lower()
            if "down" in txt or "error" in txt or "0" in txt:
                self.title = self.status_problem
            else:
                self.title = self.status_ok
        except Exception:
            self.title = self.status_problem

    @rumps.clicked("Forcer vérification")
    def force_check(self, _):
        self.check_status()

    @rumps.clicked("Quitter")
    def quit_app(self, _):
        rumps.quit_application()

if __name__ == "__main__":
    KumaTray().run()

