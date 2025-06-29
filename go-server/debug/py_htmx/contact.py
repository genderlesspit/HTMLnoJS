# contact.py - Fixed handlers
def htmx_send(request):
    """
    Handle contact form submission
    Validates and processes contact form data
    """
    try:
        name = request.get('name', '').strip()
        email = request.get('email', '').strip()
        message = request.get('message', '').strip()

        if not name:
            return '<div class="alert alert-error"><strong>Error:</strong> Name is required!</div>'
        if not email:
            return '<div class="alert alert-error"><strong>Error:</strong> Email is required!</div>'
        if not message:
            return '<div class="alert alert-error"><strong>Error:</strong> Message is required!</div>'
        if '@' not in email:
            return '<div class="alert alert-error"><strong>Error:</strong> Please enter a valid email!</div>'

        return f'''
        <div class="alert alert-success">
            <strong>Thank you, {name}!</strong><br>
            Your message has been received. We'll get back to you at {email} soon!<br>
            <small>✅ Message sent successfully via HTMLnoJS</small>
        </div>
        '''
    except Exception as e:
        return f'''
        <div class="alert alert-error">
            <strong>Error:</strong> {str(e)}
        </div>
        '''