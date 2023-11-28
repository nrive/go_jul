function pad(num) {
    let s = String(num);
    while (s.length < 2) { s = "0" + s; }
    return s;
}
let date = new Date()
let mon = date.getMonth()
let dow = date.getDay();
let day = date.getDate();
let year = date.getFullYear();
let hour = date.getHours();
let min = date.getMinutes();
if (mon != 11) {
    day = 1
    mon = 11;
} else if (day > 24) {
    day = 1;
    mon = 11;
    year = year + 1;
} else if (dow == 5) {
    day = day + 3;
} else if (dow == 6) {
    day = day + 2;
} else if ((hour == 12) && (min > 30) || (hour > 13)) {
    day = day + 1;
}
// let countDownDate = new Date(year, mon, day, 12, 30).getTime();
let countDownDate = new Date(year, mon, day, 12, 30).getTime();
let x = setInterval(function () {
    let now = new Date().getTime();
    let remainingTotal = Math.floor((countDownDate - now) / 1000);
    let hours = pad(Math.floor(Math.abs(remainingTotal / 3600)));
    remaining = (remainingTotal % 3600);
    let minutes = pad(Math.floor(Math.abs(remaining / 60)));
    remaining = (remaining % 60);
    seconds = pad(Math.floor(Math.abs(remaining)));
    let santa = String.fromCodePoint(0x1F385);
    if (remainingTotal < 0) {
        // document.getElementById("countdown").innerHTML = santa + "+ " + hours + ":" + minutes + ":" + seconds;
        document.getElementById("countdown").innerHTML = santa + "+ " + hours + "h " + minutes + "m";
    } else {
        // document.getElementById("countdown").innerHTML = santa + "- " + hours + ":" + minutes + ":" + seconds;
        document.getElementById("countdown").innerHTML = santa + "- " + hours + "h " + minutes + "m";
    }
    // if (remaining < 0) {
    //     clearInterval(x);
    //     document.getElementById("countdown").innerHTML = santa + "- 00:00:00";
    // }
}, 1000);
