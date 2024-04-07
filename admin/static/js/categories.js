async function deleteCategory(event,categoryId){
    event.preventDefault();
    await fetch(`/category?id=${categoryId}`,{method:"DELETE"}).then((response)=>{
        console.log("Successfully deleted:",response)
        window.location.assign("/categories")
    }).catch(error=>{
        console.error("error occurred:",error)
    })
}

async function createCategory(event){
    event.preventDefault();
    const formData = new FormData(event.target);
    const formProps = Object.fromEntries(formData);
    formProps["description"]=tinymce.get("description").getContent();
    console.log(formProps);
    await fetch(`/category`,{method:"POST",body:JSON.stringify(formProps)}).then((response)=>{
        console.log("Successfully deleted:",response)
        window.location.assign("/categories")
    }).catch(error=>{
        console.error("error occurred:",error)
    })
}

async function updateCategory(event,categoryId){
    event.preventDefault();
    const formData = new FormData(event.target);
    const formProps = Object.fromEntries(formData);
    formProps["description"]=tinymce.get("description").getContent();
    console.log(formProps);
    await fetch(`/category?id=${categoryId}`,{method:"PUT",body:JSON.stringify(formProps)}).then((response)=>{
        console.log("Successfully deleted:",response)
        window.location.assign("/categories")
    }).catch(error=>{
        console.error("error occurred:",error)
    })
}
