async function deleteOrder(orderId){
    await fetch(`/order?id=${orderId}`,{method:"DELETE"}).then((response)=>{
        console.log("Successfully deleted:",response)
        window.location.assign("/orders")
    }).catch(error=>{
        console.error("error occurred:",error)
    })
}

async function createOrder(event){
    event.preventDefault();
    const formData = new FormData(event.target);
    const formProps = Object.fromEntries(formData);
    console.log(formProps);
    await fetch(`/order`,{method:"POST",body:JSON.stringify(formProps)}).then((response)=>{
        console.log("Successfully created:",response)
        window.location.assign("/orders")
    }).catch(error=>{
        console.error("error occurred:",error)
    })
}

async function updateorder(event,orderId){
    event.preventDefault();
    const formData = new FormData(event.target);
    const formProps = Object.fromEntries(formData);
    console.log(formProps);
    await fetch(`/order?id=${orderId}`,{method:"PUT",body:JSON.stringify(formProps)}).then((response)=>{
        console.log("Successfully updated:",response)
        window.location.assign("/categories")
    }).catch(error=>{
        console.error("error occurred:",error)
    })
}

async function updateAddress(event){
    event.preventDefault();
    await fetch('/addresses?userId='+event.target.value).then(async (response)=>{
        const addresses=await response.json();
        var options=[];
        for (let i = 0; i < addresses.addresses.length; i++) {
            var elem=document.createElement("option");
            elem.value=addresses.addresses[i]['id']
            elem.text=`${addresses.addresses[i]['houseArea']||''},${addresses.addresses[i]['streetName']||''},${addresses.addresses[i]['city']||''},${addresses.addresses[i]['state']||''}-${addresses.addresses[i]['pincode']||''}`
            options.push(elem);
          }
        document.getElementById("address").replaceChildren(...options);
    }).catch(error=>{
        console.log("Encountered Error in Update Address"+error);
        document.getElementById("address").replaceChildren(...[]);
    })
}
