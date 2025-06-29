# Contact form handlers

def htmx_send(request):
    """
    @rate_limit(5)
    Handle contact form submission
    Validates and processes contact form data
    """
    name = request.get('name', '')
    email = request.get('email', '')
    message = request.get('message', '')

    if not all([name, email, message]):
        return '''
        <div class="alert alert-error">
            <strong>Error:</strong> All fields are required!
        </div>
        '''

    return f'''
    <div class="alert alert-success">
        <h3>Thank you, {name}!</h3>
        <p>Your message has been received. We'll get back to you at <strong>{email}</strong> soon!</p>
        <small>✅ Message sent successfully via HTMLnoJS</small>
    </div>
    '''

def htmx_validate_email(request):
    """
    Real-time email validation
    """
    email = request.get('email', '')

    if '@' not in email or '.' not in email:
        return '<small class="error">Please enter a valid email address</small>'

    return '<small class="success">✓ Valid email address</small>'