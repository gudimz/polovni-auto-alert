// Function to wait for an element to be available in the DOM
function waitForElement(selector, callback) {
    console.log('Waiting for element:', selector);
    const element = document.querySelector(selector);
    if (element) {
        console.log('Element found:', element);
        callback(element);
    } else {
        const observer = new MutationObserver((mutations, me) => {
            const element = document.querySelector(selector);
            if (element) {
                console.log('Element found via mutation observer:', element);
                callback(element);
                me.disconnect();
            }
        });
        observer.observe(document, {
            childList: true,
            subtree: true
        });
    }
}

// Function to get the list of chassis types
function getChassisTypes() {
    const chassisTypes = {};
    document.querySelectorAll('#chassis option').forEach(option => {
        if (option.value) {
            chassisTypes[option.textContent.trim()] = option.value;
        }
    });
    console.log(JSON.stringify(chassisTypes, null, 2));
}

// Start the process
waitForElement('#chassis', getChassisTypes);
