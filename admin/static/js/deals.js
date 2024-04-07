const createdeal = async (e) => {
    e.preventDefault()
   const formProps = validateDealForm(e.target)
    await fetch("/deal", {
        method:"POST", body: JSON.stringify(formProps)
    }).then(async (resp)=>{
        console.log(await resp.json())
        // window.location.assign("/deals")
    }).catch((e)=>{console.log(e)})
}

const updatedeal = async (e, dealID) => {
    e.preventDefault()
   const formProps = validateDealForm(e.target)
    await fetch("/deal?dealID="+dealID, {
        method:"PUT", body: JSON.stringify(formProps)
    }).then(async (resp)=>{
        console.log(await resp.json())
        // window.location.assign("/deals")
    }).catch((e)=>{console.log(e)})
}


const validateDealForm = (target) => {
    const formData = new FormData(target);
    const formProps = Object.fromEntries(formData);
    formProps.startsAt += ":05Z"
    formProps.endsAt += ":05Z"
    formProps.dealPrice = Number(formProps.dealPrice)
    formProps.adminPrice = Number(formProps.adminPrice)
    formProps.memberPrice = Number(formProps.memberPrice)
    formProps.quantity = Number(formProps.quantity)
    formProps.teamCapacity = Number(formProps.teamCapacity)
    formProps.isTeamDeal = formProps.isTeamDeal === "on" ? true : false
    return formProps
}

const deletedeal = async (e, dealID) =>{
    e.preventDefault()
    await fetch("/deal?dealID="+dealID, {
        method:"DELETE"
    }).then(async (resp)=>{
        console.log(await resp.json())
        // window.location.assign("/deals")
    }).catch((e)=>{console.log(e)})
}