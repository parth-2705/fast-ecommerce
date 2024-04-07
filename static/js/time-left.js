
const subtractSeconds = (date, seconds) => {
    let result = new Date(date);
    result.setSeconds(result.getSeconds() - seconds);
    return result;
};
  

function timeLeft(endsAt){
    const date1 = new Date(Date.now());
    const date2 = new Date(endsAt);
    const diffTime =date2 - date1;
    let diffDays = parseInt(diffTime / 86400000); // days
    let diffHours = parseInt((diffTime % 86400000) / 3600000); // hours
    let diffMinutes =parseInt(((diffTime % 86400000) % 3600000) / 60000); // minutes
    let diffSeconds =parseInt((((diffTime % 86400000) % 3600000)%60000) / 60000); // seconds
    return diffDays+":"+diffHours+":"+diffMinutes+":"+diffSeconds
}

function TimeLeft(endingAt) {
    const startAt = new Date();
    const endsAt = new Date(endingAt)
    const days = endsAt - startAt;
    const hours = endsAt - new Date(startAt.getTime() + parseInt(days / (24 * 60 * 60 * 1000)) * 24 * 60 * 60 * 1000);
    const minutes = endsAt - new Date(startAt.getTime() + parseInt(days / (24 * 60 * 60 * 1000)) * 24 * 60 * 60 * 1000 + parseInt(hours / (60 * 60 * 1000)) * 60 * 60 * 1000);
    const seconds = endsAt - new Date(startAt.getTime() + parseInt(days / (24 * 60 * 60 * 1000)) * 24 * 60 * 60 * 1000 + parseInt(hours / (60 * 60 * 1000)) * 60 * 60 * 1000 + parseInt(minutes / (60 * 1000)) * 60 * 1000);
    const timeLeft = parseInt(days / (24 * 60 * 60 * 1000)).toString().padStart(2, '0')+":"+parseInt(hours / (60 * 60 * 1000)).toString().padStart(2, '0')+":"+parseInt(minutes / (60 * 1000)).toString().padStart(2, '0')+":"+parseInt(seconds / 1000).toString().padStart(2, '0');
    return timeLeft;
}