document.addEventListener('DOMContentLoaded', function() {
    document.addEventListener("dblclick", event => {
        const entries = document.querySelectorAll(".entry");
        // hide them all
        entries.forEach((entry)=>{
            entry.style.display = "none";
        });

        const entriesArray = Array.from(entries);
        const shuffledArray = entriesArray.sort((a, b) => 0.5 - Math.random());
        let currentIndex = 0;

        const eWidget = document.getElementById("widget")
        eWidget.style.width = "20em";
        eWidget.style.height = "5em";

        const it = setInterval(()=>{
            const current = shuffledArray[currentIndex];
            current.style.display = "none";
            currentIndex++;
            if  (currentIndex === shuffledArray.length) {
                clearInterval(it);
                alert("Finished");
                entries.forEach((entry)=>{
                    entry.style.display = "block";
                });
                eWidget.style.display = "none";
                return;
            }
            const nextCurrent = shuffledArray[currentIndex];
            eWidget.style.display = "block";
            const word  = nextCurrent.attributes["d:title"].textContent;
            eWidget.innerHTML = "<h1>" + word + "</h1>";
            setTimeout(()=>{
                eWidget.style.display = "none";
                nextCurrent.style.display = "block";

            }, 2000);
        },5000);

    })
});


