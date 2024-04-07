function back(){
  const back = document.referrer
  console.log("document.referrer:",document.referrer)
  if(back==""){
    window.location.assign("/")
  }else if(back.includes("/otp/submit") || back.includes("/order/success")){
    window.location.assign("/")
  }else{
    history.back()
  }  
}

function targettedBack(path) {
  window.location.href = path
}