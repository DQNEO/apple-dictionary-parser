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

        const it = setInterval(()=>{
            // hide them all
            entries.forEach((entry)=>{
                entry.style.display = "none";
            });

            const entry = shuffledArray[currentIndex++];
            entry.style.display = "block";
            if  (currentIndex === shuffledArray.length) {
                clearInterval(it);
                alert("Finished");
            }
        },3000);

    })
});


