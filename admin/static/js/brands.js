async function deleteBrand(event,brandId){
    event.preventDefault();
    await fetch(`/brand?id=${brandId}`,{method:"DELETE"}).then((response)=>{
        console.log("Successfully deleted:",response)
        window.location.assign("/brands")
    }).catch(error=>{
        console.error("error occurred:",error)
    })
}

function getFeatures(){
    var featureArr=[]
    featureArr.push(document.getElementById('feature_0').value)
    featureArr.push(document.getElementById('feature_1').value)
    featureArr.push(document.getElementById('feature_2').value)
    return featureArr
}

async function createBrand(event){
    event.preventDefault();
    const formData = new FormData(event.target);
    const formProps = Object.fromEntries(formData);
    formProps["description"]=tinymce.get("description").getContent();
    console.log(formProps);
    formProps["features"]=getFeatures()
    await fetch(`/brand`,{method:"POST",body:JSON.stringify(formProps)}).then((response)=>{
        console.log("Successfully created:",response)
        window.location.assign("/brands")
    }).catch(error=>{
        console.error("error occurred:",error)
    })
}

async function updateBrand(event,brandId){
    event.preventDefault();
    const formData = new FormData(event.target);
    const formProps = Object.fromEntries(formData);
    formProps["description"]=tinymce.get("description").getContent();
    console.log(formProps);
    formProps["features"]=getFeatures()
    await fetch(`/brand?id=${brandId}`,{method:"PUT",body:JSON.stringify(formProps)}).then((response)=>{
        console.log("Successfully updated:",response)
        window.location.assign("/brands")
    }).catch(error=>{
        console.error("error occurred:",error)
    })
}
