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

// Function to get the list of models for the selected make
function getModels(makeElement, callback) {
    console.log('Selecting make element:', makeElement);
    makeElement.selected = true;
    const event = new Event('change', { bubbles: true });
    makeElement.dispatchEvent(event);

    setTimeout(() => {
        const models = [];
        document.querySelectorAll('#model option').forEach(model => {
            if (model.value) {
                models.push(model.value);
            }
        });
        console.log('Models found:', models);
        callback(models);
    }, 2000); // Increase the delay to 2 seconds
}

// Main function to get all makes and models
function getMakesAndModels() {
    const makesAndModels = {};
    const makeElements = document.querySelectorAll('#brand option');

    let index = 0;

    function processNextMake() {
        if (index >= makeElements.length) {
            console.log(JSON.stringify(makesAndModels, null, 2));
            return;
        }

        const makeElement = makeElements[index];
        const makeId = makeElement.value;
        if (makeId) {
            console.log(`Processing make ID: ${makeId}`); // Log the start of make processing
            getModels(makeElement, models => {
                makesAndModels[makeId] = models;
                index++;
                processNextMake();
            });
        } else {
            index++;
            processNextMake();
        }
    }

    processNextMake();
}

// Start the process
waitForElement('#brand', getMakesAndModels);
