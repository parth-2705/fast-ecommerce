document.addEventListener("DOMContentLoaded", () => {
    document.querySelector("#place-order").addEventListener("click", ()=>{
        // when clicked on place order make post request to the url that is currently in the browser
        fetch(window.location.href, {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                "order": "place"
            })
        })
    })        
});