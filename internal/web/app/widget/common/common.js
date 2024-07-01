
const BadId = -1;

function getDocElByClass(n) {
    return getElByClass(document, n);
}

function getElByClass(el, name) {
    return el.getElementsByClassName(name)[0];
}

// randomInt returns a random int in [0..(n-1)].
function randomInt(n) {
    return Math.floor(Math.random() * n)
}

// addCheckMark adds a <span> containing a ✔ (check mark) to the given element.
function addCheckMark(el) {
    let c = document.createElement('span');
    c.setAttribute('class', 'checkMarkUnicode');
    // https://www.compart.com/en/unicode/U+2714
    c.innerHTML = "&#x2714;";
    el.appendChild(c);
}
