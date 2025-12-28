// pkit web interface JavaScript

// Show message to user
function showMessage(text, type = 'info') {
    const messageDiv = document.createElement('div');
    messageDiv.className = `message message-${type}`;
    messageDiv.textContent = text;

    const main = document.querySelector('.main .container');
    if (main) {
        main.insertBefore(messageDiv, main.firstChild);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            messageDiv.remove();
        }, 5000);
    }
}

// Toggle bookmark
function toggleBookmark(promptID, button) {
    // Disable button during request
    button.disabled = true;
    button.classList.add('loading');

    // Send request
    fetch('/api/bookmarks', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            prompt_id: promptID
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            // Update button state
            if (data.bookmarked) {
                button.textContent = 'Remove Bookmark';
                button.classList.remove('btn-primary');
                button.classList.add('btn-danger');
            } else {
                button.textContent = 'Bookmark';
                button.classList.remove('btn-danger');
                button.classList.add('btn-primary');
            }

            // Show success message
            showMessage(data.message, 'success');
        } else {
            showMessage(data.error || 'Failed to toggle bookmark', 'error');
        }
    })
    .catch(error => {
        console.error('Bookmark toggle error:', error);
        showMessage('Network error. Please try again.', 'error');
    })
    .finally(() => {
        button.disabled = false;
        button.classList.remove('loading');
    });
}

// Update tags
function updateTags(promptID, tagsString) {
    // Parse tags from comma-separated string
    const tags = tagsString
        .split(',')
        .map(t => t.trim().toLowerCase())
        .filter(t => t.length > 0);

    // Send request
    fetch('/api/tags', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            prompt_id: promptID,
            tags: tags
        })
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            // Update current tags display
            const currentTagsDiv = document.getElementById('current-tags');
            if (data.tags && data.tags.length > 0) {
                currentTagsDiv.innerHTML = data.tags
                    .map(tag => `<span class="tag">${tag}</span>`)
                    .join('');
            } else {
                currentTagsDiv.innerHTML = '<p style="color: #7f8c8d; font-size: 14px;">No tags yet. Add some below!</p>';
            }

            // Show success message
            showMessage(data.message, 'success');
        } else {
            showMessage(data.error || 'Failed to update tags', 'error');
        }
    })
    .catch(error => {
        console.error('Tag update error:', error);
        showMessage('Network error. Please try again.', 'error');
    });
}

// Copy to clipboard
function copyToClipboard(text, button) {
    // Try modern Clipboard API first
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text)
            .then(() => {
                showCopySuccess(button);
            })
            .catch(err => {
                console.error('Clipboard API failed, trying fallback:', err);
                copyWithExecCommand(text, button);
            });
    } else {
        // Fallback for older browsers
        copyWithExecCommand(text, button);
    }
}

// Fallback copy method using execCommand
function copyWithExecCommand(text, button) {
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.left = '-9999px';
    textarea.style.top = '0';
    document.body.appendChild(textarea);

    try {
        textarea.select();
        textarea.setSelectionRange(0, textarea.value.length);

        const successful = document.execCommand('copy');
        if (successful) {
            showCopySuccess(button);
        } else {
            showMessage('Failed to copy to clipboard', 'error');
        }
    } catch (err) {
        console.error('execCommand copy failed:', err);
        showMessage('Copy to clipboard not supported', 'error');
    } finally {
        document.body.removeChild(textarea);
    }
}

// Show copy success feedback
function showCopySuccess(button) {
    const originalText = button.textContent;

    button.textContent = '✓ Copied!';
    button.classList.add('btn-success');
    button.classList.remove('btn-secondary');

    showMessage('Copied to clipboard', 'success');

    // Restore button after 2 seconds
    setTimeout(() => {
        button.textContent = originalText;
        button.classList.remove('btn-success');
        button.classList.add('btn-secondary');
    }, 2000);
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    console.log('pkit web interface loaded');

    // Add Enter key handler for tag input
    const tagInput = document.getElementById('tag-input');
    if (tagInput) {
        tagInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                // Find the prompt ID from the save button
                const saveButton = tagInput.parentElement.querySelector('button');
                if (saveButton) {
                    saveButton.click();
                }
            }
        });
    }
});
