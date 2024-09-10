class App {
    showAnswer() {
        this.eWidget.style.display = "none";
        const entry = this.entries[this.currentIndex];
        entry.style.display = "block";
    }
    showQuestion() {
        const entry = this.entries[this.currentIndex];
        while (this.eWidget.firstChild) {
            this.eWidget.removeChild(this.eWidget.firstChild);
        }
        const eHG = entry.getElementsByClassName("hg")[0];
        const clonedHG = eHG.cloneNode(true);
        this.eWidget.appendChild(clonedHG);
        this.eWidget.style.display = "block";
    }
    init() {
        this.running = true;
        document.getElementById("widget-starter").style.display = "none";
        const entries = document.querySelectorAll(".entry");
        // hide them all
        entries.forEach((entry) => {
            entry.style.display = "none";
        });

        this.entries = Array.from(entries).sort((a, b) => 0.5 - Math.random());
        this.currentIndex = 0;

        this.eWidget = document.getElementById("widget")
        this.eWidget.className = "entry";
        this.eWidget.style.borderTop = "1px solid #909090";
        this.eWidget.innerHTML = `<h1>Quiz</h1>`
        this.eWidget.style.display = "block";
        this.status = "";

        document.onclick = () => {
            switch (this.status) {
                case "":
                    this.showQuestion();
                    this.status = "asking";
                    break;
                case "asking":
                    this.showAnswer();
                    this.status = "answering";
                    break;
                case "answering":
                    const cur = this.entries[this.currentIndex];
                    cur.style.display = "none";
                    this.currentIndex++;
                    if (this.currentIndex === this.entries.length) {
                        alert("Finished");
                        this.status = "";
                        this.finish();
                        return;
                    }
                    this.showQuestion();
                    this.status = "asking";
                    break;
            }
        };
    }
    finish() {
        this.entries.forEach((entry) => {
            entry.style.display = "block";
        });
        this.eWidget.style.display = "none";
        this.currentIndex = 0;
        this.running = false;
    }
}

const app = new App();

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('widget-starter-button').addEventListener("click", ()=>{
        if (app.running) {
            return;
        }
        app.init();
    });
});

